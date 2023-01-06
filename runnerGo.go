package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"

	// _ "net/http/pprof"

	"github.com/Apipost-Team/runnerGo/tools"
	"github.com/Apipost-Team/runnerGo/worker"
	"golang.org/x/net/websocket"
)

var urlsBlacklist = []string{".apis.cloud", ".apipost.cn", ".apipost.com", ".apipost.net", ".runnergo.com", ".runnergo.cn", ".runnergo.net"}

func main() {
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
					continue
				}
				//取消压测target_id, cancel:xxxxxxxxxx
				if strings.HasPrefix(body, "cancel:") {
					target_id := body[7:]
					control, ok := controlMap[target_id]
					if !ok {
						msg := `{"code":501, "message":"任务不存在，无需取消", "data":{}}`
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
						msg := `{"code":501, "message":"任务不存在", "data":{"Target_id":"` + target_id + `"}}`
						sendChan <- msg
						continue
					}

					jsonRes, err := json.Marshal(control)
					if err != nil {
						log.Println(err.Error())
						continue
					}

					msg := `{"code":200, "message":"success", "data":` + string(jsonRes) + `}`
					sendChan <- msg
					continue
				}

				if strings.HasPrefix(body, "quit") {
					//退出
					os.Exit(0)
				}

				var bodyStruct worker.RawWorkData

				// 解析 har 结构
				json.Unmarshal([]byte(string(body)), &bodyStruct)
				control := tools.ControlData{
					C:          bodyStruct.C,
					N:          bodyStruct.N,
					Total:      bodyStruct.C * bodyStruct.N,
					Target_id:  bodyStruct.Target_id,
					MaxRunTime: bodyStruct.MaxRunTime, //10分钟
					IsCancel:   false,
					IsRunning:  false,
					TimeOut:    20, //超时时间
				}

				if control.MaxRunTime < 1 {
					control.MaxRunTime = 600 //10分钟
				}

				if control.TimeOut > control.MaxRunTime {
					control.TimeOut = control.MaxRunTime //超时时间不能超过总时间
				}

				log.Println(control)

				if control.Total <= 0 {
					msg := `{"code":501, "message":"并发数或者循环次数至少为1", "data":{"Target_id":"` + control.Target_id + `"}}`
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
