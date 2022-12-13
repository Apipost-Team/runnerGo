package worker

import (
	"context"
	"encoding/json"
	"os/exec"
	"runtime"
	"sync"
	"time"

	runnerHttp "github.com/Apipost-Team/runnerGo/http"
	"github.com/Apipost-Team/runnerGo/summary"
	"github.com/Apipost-Team/runnerGo/tools"
	"golang.org/x/net/websocket"
)

type InputData struct {
	C         int    `json:"c"`
	N         int    `json:"n"`
	Target_id string `json:"target_id"`
	Data      runnerHttp.HarRequestType
}

// 添加任务
func AddTask(control tools.ControlData, data runnerHttp.HarRequestType, urlChanel chan runnerHttp.HarRequestType, ctx context.Context) {
	for i := 0; i < control.Total; i++ {
		urlChanel <- data
	}

	for {
		select {
		case <-ctx.Done():
			close(urlChanel)
			return
		default:
			time.Sleep(time.Duration(50) * time.Millisecond)
			if len(urlChanel) == 0 {
				close(urlChanel)
				return
			}
		}
	}
}

// 执行请求任务
func worker(urlChanel chan runnerHttp.HarRequestType, ws *websocket.Conn, ctx context.Context) {
	for { // for 循环逐个执行 URL
		select {
		case <-ctx.Done():
			return
		default:
			data, ok := <-urlChanel
			if !ok {
				return
			}
			summary.ResChanel <- runnerHttp.Do(data, ws)
		}
	}
}

// 开始任务
func StartWork(control tools.ControlData, data runnerHttp.HarRequestType, ws *websocket.Conn) {
	var rwg sync.WaitGroup
	var urlChanel = make(chan runnerHttp.HarRequestType)
	summary.ResChanel = make(chan summary.Res)

	ctx, _ := context.WithTimeout(context.Background(), time.Second*time.Duration(control.MaxRunTime))

	rwg.Add(1)

	// 添加任务
	go AddTask(control, data, urlChanel, ctx)

	// 并发消费 请求
	for i := 0; i < control.C; i++ {
		go func() {
			worker(urlChanel, ws, ctx)
		}()
	}

	// 处理数据
	go func() {
		res := summary.HandleRes(control, ctx)
		jsonRes, err := json.Marshal(res)

		if err != nil {
			summary.SendResult(string(err.Error()), 503, ws)
		} else {
			summary.SendResult(string(jsonRes), 200, ws)
		}

		rwg.Done()
	}()

	rwg.Wait()
}

// 用默认浏览器打开指定URL
func OpenUrl(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"

	}

	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}
