package worker

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"sync/atomic"
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
	//control  初始化
	control.StartTime = int(tools.GetNowUnixNano())
	control.EndTime = control.StartTime
	control.IsRunning = true //设置为启动
	control.IsCancel = false //设置为取消
	control.WorkCnt = 0

	var urlChanel = make(chan runnerHttp.HarRequestType, 8) //8个url任务列表,带缓冲
	var resultChanel = make(chan summary.Res, 16)           //返回结果列表，带缓冲

	defer func() {
		control.IsRunning = false
		//关闭所有通道，强制释放协程序,统一结束时候是否
		close(urlChanel) //关闭url任务发送，清理进程
		//清理resultChanel
		if control.WorkCnt > 2 {
			preWorkCnt := control.WorkCnt

		OutClean:
			for control.WorkCnt > 2 {
				select {
				case res := <-resultChanel: //结束进程清理，唯一消耗方，不会阻塞
					_ = res //忽略错误
					continue

				case <-time.After(100 * time.Millisecond):
					//定时检查work是否在减少，没有就直接退出
					if control.WorkCnt < preWorkCnt {
						preWorkCnt = control.WorkCnt //记录现在的数量
						continue
					}
					break OutClean
				}
			}
		}
		close(resultChanel)
		log.Println("action main quit", *control)
	}() //设置为执行完成

	defer func() {
		if r := recover(); r != nil {
			log.Println("执行任务出错,忽略退出", r)
			msg := fmt.Sprintf("\"code\":501, \"message\": \"%q\", \"data\":{\"Target_id\":\"%s\"}}", r, control.Target_id)
			sendChan <- msg
		}
	}()

	//注册取消操作
	go func(control *tools.ControlData) {
		log.Println("run timeout", control.MaxRunTime)
		timeChan := time.After(time.Second * time.Duration(control.MaxRunTime))
	OutCancel:
		for {
			select {
			case <-timeChan:
				control.IsCancel = true //设置为取消
				log.Println("action timout")
				break OutCancel
			case <-time.After(50 * time.Millisecond):
				if control.IsCancel || (!control.IsRunning) { //取消或者支持完成直接退出
					if control.IsCancel {
						log.Println("action cancel")
					} else {
						log.Println("action end")
					}

					break OutCancel
				}
			}
		}
	}(control)

	//设置并发任务消费,需要连接池
	tr := &http.Transport{
		//MaxConnsPerHost: 2000, //限定2k连接
		MaxConnsPerHost: 0,
		IdleConnTimeout: 10 * time.Second,
		// MaxIdleConnsPerHost: control.C + 128,
		DisableKeepAlives:  false,
		DisableCompression: false,
	}

	for i := 0; i < control.C; i++ {
		go doWork(control, tr, i, urlChanel, resultChanel)
	}

	//添加任务呢
	go func(urlChanel chan<- runnerHttp.HarRequestType, control *tools.ControlData) {
		defer func() {
			if r := recover(); r != nil {
				log.Println("error add task", r)
			}
		}()

		for i := 0; i < control.Total; i++ {
			if control.IsCancel {
				log.Println("action add task quit")
				break
			}
			urlChanel <- data
		}
	}(urlChanel, control)

	//统计结果
	res := summary.HandleRes(control, resultChanel)
	jsonRes, err := json.Marshal(res)

	var msg string
	if err != nil {
		msg = `{"code":501, "message":"` + string(err.Error()) + `", "data":{"Target_id":"` + control.Target_id + `"}}`
	} else {
		msg = `{"code":200, "message":"success", "data":` + string(jsonRes) + `}`
	}

	sendChan <- msg
}

func doWork(control *tools.ControlData, tr *http.Transport, i int, urlChanel <-chan runnerHttp.HarRequestType, resultChanel chan<- summary.Res) {
	atomic.AddInt32(&(control.WorkCnt), 1) //进程数加1
	defer func() {
		atomic.AddInt32(&(control.WorkCnt), -1) //工作线程减1
	}()
	defer func() {
		if r := recover(); r != nil {
			log.Println("work error", i, r)
		}
	}()

	//初始化 httpclient
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Duration(control.TimeOut) * time.Second,
	}

	for { // for 循环逐个执行 URL
		if control.IsCancel {
			break
		} else {
			data, ok := <-urlChanel
			if !ok {
				break
			}
			resultChanel <- runnerHttp.Do(client, data)
		}
	}
}
