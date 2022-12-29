package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	runnerHttp "github.com/Apipost-Team/runnerGo/http"
	"github.com/Apipost-Team/runnerGo/summary"
	"github.com/Apipost-Team/runnerGo/tools"
)

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

func Process(control *tools.ControlData, data runnerHttp.HarRequestType, sendChan chan<- string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("执行任务出错,忽略退出", r)
			msg := fmt.Sprintf("\"code\":501, \"message\": \"%q\", \"data\":{}}", r)
			sendChan <- msg
		}
	}()
	//control  初始化
	control.StartTime = int(tools.GetNowUnixNano())
	control.EndTime = control.StartTime
	control.IsRunning = true //设置为启动

	defer func() { control.IsRunning = false }() //设置为执行完成

	var urlChanel = make(chan runnerHttp.HarRequestType, 10) //url任务列表,带缓冲
	var resultChanel = make(chan summary.Res, 10)            //返回结果列表，带缓冲

	ctx, cancelFun := context.WithCancel(context.Background()) //主动取消
	//注册取消操作
	go func(cancelFun context.CancelFunc) {
		fmt.Println("超时时间", control.MaxRunTime)
		timeChan := time.After(time.Second * time.Duration(control.MaxRunTime))
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, syscall.SIGUSR2)
		for {
			select {
			case <-ctx.Done():
				close(urlChanel) //阻止发送数据
				fmt.Println("任务结束")
				return
			case <-timeChan:
				fmt.Println("超时关闭")
				close(urlChanel) //阻止发送数据
				return
			case <-signalChan:
				if control.IsCancel {
					fmt.Println("主动关闭")
					cancelFun() //取消所有任务
					return
				}
			}
		}
	}(cancelFun)
	defer cancelFun() //主动取消

	//设置并发任务消费
	for i := 0; i < control.C; i++ {
		go doWork(*control, urlChanel, resultChanel, ctx)
	}

	//添加任务呢
	go func(urlChanel chan<- runnerHttp.HarRequestType, ctx context.Context) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("添加任务失败,忽略退出", r)
			}
		}()

		doneChan := ctx.Done()
		for i := 0; i < control.Total; i++ {
			select {
			case <-doneChan:
				fmt.Println("关闭任务发送")
				close(urlChanel)
				return
			default:
				urlChanel <- data
			}
		}
	}(urlChanel, ctx)

	//统计结果
	res := summary.HandleRes(*control, resultChanel, ctx)
	jsonRes, err := json.Marshal(res)

	var msg string
	if err != nil {
		msg = `{"code":501, "message":"` + string(err.Error()) + `", "data":{}}`
	} else {
		msg = `{"code":200, "message":"success", "data":` + string(jsonRes) + `}`
	}

	sendChan <- msg

}

func doWork(control tools.ControlData, urlChanel <-chan runnerHttp.HarRequestType, resultChanel chan<- summary.Res, ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("执行任务出错,忽略退出", r)
		}
	}()
	doneChan := ctx.Done()

	//初始化 httpclient
	client := &http.Client{
		// Transport: &http.Transport{
		// 	// MaxConnsPerHost:     control.C + 1,
		// 	// MaxIdleConnsPerHost: control.C + 128,
		// 	DisableKeepAlives:  false,
		// 	DisableCompression: false,
		// },
		Timeout: time.Duration(control.TimeOut) * time.Second,
	}

	for { // for 循环逐个执行 URL
		select {
		case <-doneChan:
			return
		default:
			data, ok := <-urlChanel
			if !ok {
				return
			}
			resultChanel <- runnerHttp.Do(client, data)
		}
	}
}
