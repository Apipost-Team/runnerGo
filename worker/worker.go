package worker

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"sync/atomic"
	"time"

	runnerHttp "github.com/Apipost-Team/runnerGo/http"
	"github.com/Apipost-Team/runnerGo/summary"
	"github.com/Apipost-Team/runnerGo/tools"
)

type NullWriter int

func (NullWriter) Write([]byte) (int, error) { return 0, nil }

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
	control.WorkCnt = 0      //设置工作进程为0

	var urlChanel = make(chan runnerHttp.HarRequestType, 8) //8个url任务列表,带缓冲
	var resultChanel = make(chan summary.Res, 16)           //返回结果列表，带缓冲
	var rLog *log.Logger                                    //日志文件

	//设置并发任务消费,需要连接池
	defaultCipherSuites := []uint16{0xc02f, 0xc030, 0xc02b, 0xc02c, 0xcca8, 0xcca9, 0xc013, 0xc009,
		0xc014, 0xc00a, 0x009c, 0x009d, 0x002f, 0x0035, 0xc012, 0x000a}
	tr := &http.Transport{
		MaxConnsPerHost:    0,
		IdleConnTimeout:    20 * time.Second,
		DisableKeepAlives:  false,
		DisableCompression: false,
		TLSClientConfig: &tls.Config{
			CipherSuites: append(defaultCipherSuites[8:], defaultCipherSuites[:8]...),
		}, //绕过特征检测
		TLSHandshakeTimeout: 10 * time.Second,
	}

	//程序错误处理
	defer func() {
		if r := recover(); r != nil {
			log.Println("执行任务出错,忽略退出", r)
			msg := fmt.Sprintf("\"code\":501, \"message\": \"%q\", \"data\":{\"Target_id\":\"%s\"}}", r, control.Target_id)
			sendChan <- msg
		}
	}()

	//工作接收清理工作
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

	//判断是否要创建文本日志
	if control.LogType > 0 {
		f, err := os.CreateTemp("", "runnergo_")
		if err != nil {
			//出错了，无法开启日志
			rLog.SetFlags(0)                //屏蔽所有日志
			rLog.SetOutput(new(NullWriter)) //设置空接口
		} else {
			defer f.Close() //关闭日志
			rLog = log.New(f, "", log.Ltime)
			log.Printf("log file:%s", f.Name())
		}
	}

	//运行是控制，取消，超时，控制进程数量，定时汇报
	go func(control *tools.ControlData) {
		defer func() {
			if r := recover(); r != nil {
				log.Println("error control", r)
			}
		}()

		timeChan := time.After(time.Second * time.Duration(control.MaxRunTime)) //超时设置
		checkCnt := 0                                                           //检查次数
		checkInterval := 50                                                     //检查间隔
		reportInterval := 0                                                     //反馈间隔
		if control.ReportTime > 0 {
			reportInterval = int(float64(control.ReportTime)/float64(checkInterval) + 0.5) //向上取整数
			if reportInterval < 1 {
				reportInterval = 1 //report时间最小50毫秒
			}
		}

		log.Println("run timeout", control.MaxRunTime, "reportInterval", reportInterval, "reportTime", control.ReportTime)

	OutCancel:
		for {
			select {
			case <-timeChan:
				control.IsCancel = true //设置为取消
				log.Println("action timout")
				break OutCancel
			case <-time.After(time.Duration(checkInterval) * time.Millisecond):
				checkCnt++ //检查次数+1
				//确定退出
				if control.IsCancel || (!control.IsRunning) { //取消或者支持完成直接退出
					if control.IsCancel {
						log.Println("action cancel")
					} else {
						log.Println("action end")
					}
					break OutCancel
				}

				//检查report
				if reportInterval > 0 && (checkCnt%reportInterval == 0) {
					jsonRes, err := json.Marshal(control)
					if err != nil {
						log.Println(err.Error())
						continue
					}

					msg := `{"code":202, "message":"success", "data":` + string(jsonRes) + `}`
					sendChan <- msg //发送统计信息
				}

				//检查是否需要增加进程
				if control.WorkTagetCnt > 0 {
					diffCnt := int(control.WorkTagetCnt - control.WorkCnt) //目标和实际差距
					if diffCnt > 0 {
						//增加进程
						for i := 0; i < diffCnt; i++ {
							go doWork(control, tr, i+int(control.WorkTagetCnt), urlChanel, resultChanel, rLog)
						}
						control.WorkTagetCnt = 0 //启动完成，标记
					} else if diffCnt < 0 {
						//减少进程
						diffCnt = -diffCnt

						for i := 0; i < diffCnt; i++ {
							emptyData := runnerHttp.HarRequestType{}
							emptyData.Seq = -1
							urlChanel <- emptyData
						}
						control.WorkTagetCnt = 0 //可能阻塞,标记
					}
				}

			}
		}
	}(control)

	//启动工作进程
	for i := 0; i < control.C; i++ {
		go doWork(control, tr, i, urlChanel, resultChanel, rLog)
	}

	//创建工作任务
	go func(urlChanel chan<- runnerHttp.HarRequestType, data runnerHttp.HarRequestType, control *tools.ControlData) {
		defer func() {
			if r := recover(); r != nil {
				log.Println("error add task", r)
			}
		}()

		for i := 0; i < control.Total; i++ {
			if control.IsCancel || (!control.IsRunning) {
				log.Println("action add task quit")
				break
			}

			data.Seq = i //设置请求序列
			urlChanel <- data
		}
	}(urlChanel, data, control)

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

func doWork(control *tools.ControlData, tr *http.Transport, workId int, urlChanel <-chan runnerHttp.HarRequestType, resultChanel chan<- summary.Res, rLog *log.Logger) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("work error", workId, r)
		}
	}()

	//设置数量
	atomic.AddInt32(&(control.WorkCnt), 1) //进程数加1
	defer func() {
		atomic.AddInt32(&(control.WorkCnt), -1) //工作线程减1
	}()

	//初始化 httpclient
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Duration(control.TimeOut) * time.Second,
	}

	for { // for 循环逐个执行 URL
		if control.IsCancel || (!control.IsRunning) {
			break
		} else {
			data, ok := <-urlChanel
			if !ok {
				//工作已经执行完成
				break
			}

			if data.Seq < 0 {
				//只有控制信息seq会为负数
				break
			}

			result, err := runnerHttp.Do(client, data)
			//日志记录
			if control.LogType > 0 {
				if err != nil {
					rLog.Printf("seq:%d\twork:%d\tcode:%d\tcost:%.1f\tmsg:%s", data.Seq, workId, result.Code, result.TotalUseTime, err.Error())
				} else if control.LogType < 2 {
					rLog.Printf("seq:%d\twork:%d\tcode:%d\tcost:%.1f\tmsg:%s", data.Seq, workId, result.Code, result.TotalUseTime, "ok")
				}
			}

			resultChanel <- result
		}
	}
}
