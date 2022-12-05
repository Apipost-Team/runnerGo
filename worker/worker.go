package worker

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/Apipost-Team/runnerGo/conf"
	runnerHttp "github.com/Apipost-Team/runnerGo/http"
	"github.com/Apipost-Team/runnerGo/summary"
	"golang.org/x/net/websocket"
)

type InputData struct {
	C    int `json:"c"`
	N    int `json:"n"`
	Data runnerHttp.HarRequestType
}

// 结果反馈
func SendResult(msg string, ws *websocket.Conn) {
	if err := websocket.Message.Send(ws, msg); err != nil {
		panic(err)
	}
}

// 添加任务
func AddTask(data runnerHttp.HarRequestType, urlChanel chan runnerHttp.HarRequestType) {
	for i := 0; i < conf.Conf.UrlNum; i++ {
		urlChanel <- data
	}

	for {
		time.Sleep(time.Duration(50) * time.Millisecond)
		if len(urlChanel) == 0 {
			close(urlChanel)
			return
		}
	}
}

// 执行请求任务
func worker(urlChanel chan runnerHttp.HarRequestType) {
	for { // for 循环逐个执行 URL
		data, ok := <-urlChanel
		if !ok {
			return
		}
		summary.ResChanel <- runnerHttp.Do(data)
	}
}

// 开始任务
func StartWork(data runnerHttp.HarRequestType, ws *websocket.Conn) {
	var rwg sync.WaitGroup
	var urlChanel = make(chan runnerHttp.HarRequestType)
	summary.ResChanel = make(chan summary.Res)

	rwg.Add(1)

	// 添加任务
	go AddTask(data, urlChanel)

	// 并发消费 请求
	for i := 0; i < conf.Conf.C; i++ {
		go func() {
			worker(urlChanel)
		}()
	}

	// 处理数据
	go func() {
		res := summary.HandleRes()
		jsonRes, err := json.Marshal(res)

		if err != nil {
			panic(err)
		}

		SendResult(string(jsonRes), ws)
		rwg.Done()
	}()

	rwg.Wait()
}
