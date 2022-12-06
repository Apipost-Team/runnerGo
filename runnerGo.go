package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/Apipost-Team/runnerGo/conf"
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
				fmt.Println("receive failed:", err)
				break
			}
			var bodyStruct worker.InputData

			// 解析 har 结构
			json.Unmarshal([]byte(string(body)), &bodyStruct)
			conf.Conf.C = bodyStruct.C
			conf.Conf.UrlNum = bodyStruct.C * bodyStruct.N

			// 开始时间
			conf.Conf.StartTime = int(tools.GetNowUnixNano())

			// 开始压测
			worker.StartWork(bodyStruct.Data, ws)

		}
	}))

	if err := http.ListenAndServe(":10397", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
