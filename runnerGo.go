package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"

	//_ "net/http/pprof"

	"github.com/Apipost-Team/runnerGo/tools"
	"github.com/Apipost-Team/runnerGo/worker"
	"golang.org/x/net/websocket"

	runnerHttp "github.com/Apipost-Team/runnerGo/http"
)

var urlsBlacklist = []string{".apis.cloud", ".apipost.cn", ".apipost.com", ".apipost.net", ".runnergo.com", ".runnergo.cn", ".runnergo.net"}

func main() {
	http.HandleFunc("/quit", func(w http.ResponseWriter, r *http.Request) {
		//增加退出代码
		log.Println("user quit")
		os.Exit(0)
	})
	http.HandleFunc("/send", func(w http.ResponseWriter, r *http.Request) {
		//增加代理发送功能
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println(string(body))
		var p runnerHttp.HarRequestType
		err = json.Unmarshal(body, &p)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println(p)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{}"))
	})
	http.Handle("/websocket", websocket.Handler(func(ws *websocket.Conn) {
		var sendChan = make(chan string)
		defer ws.Close()

		go func(sendChan chan<- string, ws *websocket.Conn) {
			var controlMap = make(map[string]*tools.ControlData)
			for {
				var body string
				if err := websocket.Message.Receive(ws, &body); err != nil {
					fmt.Println("read error")
					fmt.Println(err)
					for _, control := range controlMap {
						if control.IsRunning {
							control.IsCancel = true //主动设置为取消
						}
					}
					ws.Close()
					break
				}
				//fmt.Println("Received: ", body)
				if strings.HasPrefix(body, "PING") || strings.HasPrefix(body, "ping") {
					//处理ping
					continue
				}
				//取消压测target_id, cancel:xxxxxxxxxx
				if strings.HasPrefix(body, "cancel:") {
					target_id := body[7:]
					control, ok := controlMap[target_id]
					if !ok {
						msg := `{"code":501, "message":"任务不存在，无需取消", "data":{"Target_id":"` + target_id + `"}}`
						sendChan <- msg
						continue
					}

					if !control.IsRunning {
						msg := `{"code":501, "message":"任务已结束，无需终止", "data":{"Target_id":"` + target_id + `"}}`
						sendChan <- msg
						continue
					}

					control.IsCancel = true
					continue
				}

				//查询target_id执行情况, query:xxxxxxxxxx
				if strings.HasPrefix(body, "query:") {
					target_id := body[6:]
					control, ok := controlMap[target_id]
					if !ok {
						msg := `{"code":502, "message":"任务不存在", "data":{"Target_id":"` + target_id + `"}}`
						sendChan <- msg
						continue
					}

					jsonRes, err := json.Marshal(control)
					if err != nil {
						log.Println(err.Error())
						continue
					}

					msg := `{"code":201, "message":"success", "data":` + string(jsonRes) + `}`
					sendChan <- msg
					continue
				}

				if strings.HasPrefix(body, "quit") {
					//退出
					os.Exit(0)
				}

				if strings.HasPrefix(body, "setc:") {
					//setc:target_id:num  将并发数设置到指定
					tmpArr := strings.Split(body, ":")
					if len(tmpArr) != 3 {
						msg := `{"code":504, "message":"setc invalid", "data":''}`
						sendChan <- msg
						continue
					}
					target_id := tmpArr[1]
					WorkTagetCnt, err := strconv.Atoi(tmpArr[2])
					if err != nil {
						msg := `{"code":501, "message":"num is invalid", "data":{"Target_id":"` + target_id + `"}}`
						sendChan <- msg
						continue
					}
					control, ok := controlMap[target_id]
					if !ok {
						msg := `{"code":501, "message":"任务不存在，无法设置", "data":{"Target_id":"` + target_id + `"}}`
						sendChan <- msg
						continue
					}

					if !control.IsRunning {
						msg := `{"code":501, "message":"任务已结束，无法设置", "data":{"Target_id":"` + target_id + `"}}`
						sendChan <- msg
						continue
					}

					control.WorkTagetCnt = int32(WorkTagetCnt) //设置数量
					continue
				}

				//处理发送请求
				var bodyStruct worker.RawWorkData

				// 解析 har 结构
				json.Unmarshal([]byte(string(body)), &bodyStruct)
				control := tools.ControlData{
					C:            bodyStruct.C,
					N:            bodyStruct.N,
					Total:        bodyStruct.C * bodyStruct.N,
					Target_id:    bodyStruct.Target_id,
					MaxRunTime:   bodyStruct.MaxRunTime, //10分钟
					IsCancel:     false,
					IsRunning:    false,
					TimeOut:      20,                    //超时时间
					LogType:      bodyStruct.LogType,    //是否开启日志
					ReportTime:   bodyStruct.ReportTime, //是否定时汇报日志
					WorkTagetCnt: 0,                     //默认设置为0
				}

				if control.MaxRunTime < 1 {
					control.MaxRunTime = 600 //10分钟
				}

				if control.TimeOut > control.MaxRunTime {
					control.TimeOut = control.MaxRunTime //超时时间不能超过总时间
				}

				log.Println(control)

				if control.Total <= 0 && control.MaxRunTime <= 0 {
					msg := `{"code":501, "message":"并发数或者执行时间不能为0", "data":{"Target_id":"` + control.Target_id + `"}}`
					sendChan <- msg
					continue
				}

				//检查url是否被禁止
				isForbidden := false
				for i := 0; i < len(urlsBlacklist); i++ {
					if strings.Contains(strings.ToLower(bodyStruct.Data.Url), urlsBlacklist[i]) {
						isForbidden = true
						break
					}
				}

				if isForbidden {
					msg := `{"code":301, "message":"禁止请求的URL", "data":{"Target_id":"` + control.Target_id + `"}}`
					sendChan <- msg
					continue
				}

				if len(control.Target_id) > 0 {
					newControl, ok := controlMap[control.Target_id]
					if ok && newControl.IsRunning {
						msg := `{"code":301, "message":"相同的target_id任务还在执行，请等上次执行完成", "data":{"Target_id":"` + control.Target_id + `"}}`
						sendChan <- msg
						continue
					}

					//检查现在情况
					ng := runtime.NumGoroutine()
					if ng > 30000 {
						msg := `{"code":302, "message":"压测负载太高，请稍后重试", "data":{"Target_id":"` + control.Target_id + `"}}`
						sendChan <- msg
						continue
					}

					controlMap[control.Target_id] = &control
				}
				// 开始压测,改异步
				go worker.Process(&control, bodyStruct.Data, sendChan)

			}
		}(sendChan, ws)

		for {
			msg := <-sendChan
			//fmt.Println("send: ", msg)
			if err := websocket.Message.Send(ws, msg); err != nil {
				fmt.Println("write")
				fmt.Println(err)
				break
			}
		}

	}))

	if err := http.ListenAndServe(":10397", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
