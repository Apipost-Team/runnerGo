package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"text/template"

	"github.com/Apipost-Team/runnerGo/conf"
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

		for {
			var body string
			//websocket接受信息

			if err = websocket.Message.Receive(ws, &body); err != nil {
				ws.Close()
				break
			} else {
				var bodyStruct worker.InputData

				// 解析 har 结构
				json.Unmarshal([]byte(string(body)), &bodyStruct)
				conf.Conf.C = bodyStruct.C
				conf.Conf.UrlNum = bodyStruct.C * bodyStruct.N

				isForbidden := false

				for i := 0; i < len(urlsBlacklist); i++ {
					if strings.Index(strings.ToLower(bodyStruct.Data.Url), urlsBlacklist[i]) > -1 {
						isForbidden = true
						goto gotofor
					}
				}

			gotofor:
				if conf.Conf.UrlNum <= 0 {
					summary.SendResult(`并发数或者循环次数至少为1`, 501, ws)
				} else if isForbidden {
					summary.SendResult(`禁止请求的URL`, 301, ws)
				} else {
					// 开始时间
					conf.Conf.StartTime = int(tools.GetNowUnixNano())

					// 开始压测
					worker.StartWork(bodyStruct.Data, ws)
				}
			}
		}
	}))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t, _ := template.ParseFiles("websocket.html")
		t.Execute(w, nil)
	})

	go worker.OpenUrl("http://127.0.0.1:10397/")
	if err := http.ListenAndServe(":10397", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
