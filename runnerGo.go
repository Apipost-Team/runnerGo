package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/Apipost-Team/runnerGo/summary"
	"github.com/Apipost-Team/runnerGo/tools"
	"github.com/Apipost-Team/runnerGo/worker"
	"golang.org/x/net/websocket"
)

var urlsBlacklist = []string{".apis.cloud", ".apipost.cn", ".apipost.com", ".apipost.net", ".runnergo.com", ".runnergo.cn", ".runnergo.net"}

func main() {
	//接受websocket的路由地址
	http.Handle("/websocket", websocket.Handler(func(ws *websocket.Conn) {
		var err error
		var controlMap = make(map[string]*tools.ControlData)

		for {
			var body string
			//websocket接受信息

			if err = websocket.Message.Receive(ws, &body); err != nil {
				ws.Close()
				break
			} else {
				if strings.HasPrefix(body, "cancel:") {
					//取消压测target_id, cancel:xxxxxxxxxx
					log.Println(body)
					target_id := body[7:]
					control, ok := controlMap[target_id]
					if !ok {
						summary.SendResult(`{target_id} 不存在`, 501, ws)
						continue
					}

					control.IsCancel = true

					continue
				}

				if strings.HasPrefix(body, "query:") {
					//查询target_id执行情况, query:xxxxxxxxxx
					log.Println("get")
					continue
				}

				if strings.HasPrefix(body, "quit") {
					//退出
					os.Exit(0)
				}

				var bodyStruct worker.InputData

				// 解析 har 结构
				json.Unmarshal([]byte(string(body)), &bodyStruct)
				control := tools.ControlData{
					C:          bodyStruct.C,
					N:          bodyStruct.N,
					Total:      bodyStruct.C * bodyStruct.N,
					Target_id:  bodyStruct.Target_id,
					MaxRunTime: 600, //10分钟
					IsCancel:   false,
				}

				isForbidden := false

				for i := 0; i < len(urlsBlacklist); i++ {
					if strings.Index(strings.ToLower(bodyStruct.Data.Url), urlsBlacklist[i]) > -1 {
						isForbidden = true
						goto gotofor
					}
				}

			gotofor:
				if control.Total <= 0 {
					summary.SendResult(`并发数或者循环次数至少为1`, 501, ws)
				} else if isForbidden {
					summary.SendResult(`禁止请求的URL`, 301, ws)
				} else {
					// 开始时间
					control.StartTime = int(tools.GetNowUnixNano())
					control.EndTime = control.StartTime //初始化执行时间

					if len(control.Target_id) > 0 {
						controlMap[control.Target_id] = &control
					}
					// 开始压测,改异步
					worker.StartWork(&control, bodyStruct.Data, ws)
				}
			}
		}
	}))

	// http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// 	t, _ := template.ParseFiles("websocket.html")
	// 	t.Execute(w, nil)
	// })

	// go worker.OpenUrl("http://127.0.0.1:10397/")
	if err := http.ListenAndServe(":10397", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
