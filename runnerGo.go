package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/Apipost-Team/runnerGo/conf"
	"github.com/Apipost-Team/runnerGo/summary"
	"github.com/Apipost-Team/runnerGo/tools"
	"github.com/Apipost-Team/runnerGo/worker"
	"golang.org/x/net/websocket"
)

func main() {
	//接受websocket的路由地址
	http.Handle("/websocket", websocket.Handler(func(ws *websocket.Conn) {
		var err error

		for {
			var body string
			//websocket接受信息

			if err = websocket.Message.Receive(ws, &body); err != nil {
				summary.SendResult(string(err.Error()), 500, ws)
			} else {
				if strings.HasPrefix(body, "cancel") {
					//取消压测target_id, cancel:xxxxxxxxxx
					log.Println("cancel")
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
					C:         bodyStruct.C,
					N:         bodyStruct.N,
					Total:     bodyStruct.C * bodyStruct.N,
					Target_id: bodyStruct.Target_id,
				}

				if control.Total <= 0 {
					summary.SendResult(`并发数或者循环次数至少为1`, 501, ws)
				} else {
					// 开始时间
					conf.Conf.StartTime = int(tools.GetNowUnixNano())

					// 开始压测
					worker.StartWork(control, bodyStruct.Data, ws)
				}
			}
		}
	}))

	if err := http.ListenAndServe(":10397", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
