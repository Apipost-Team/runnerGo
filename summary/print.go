package summary

import (
	"fmt"
	"strconv"

	"github.com/Apipost-Team/runnerGo/conf"
	"golang.org/x/net/websocket"
)

// 结果反馈
func SendResult(msg string, code int, ws *websocket.Conn) {
	if code == 200 {
		msg = `{"code":200, "message":"success", "data":` + msg + `, "C":` + strconv.Itoa(conf.Conf.C) + `}`
	} else {
		msg = `{"code":` + strconv.Itoa(code) + `, "message":"` + msg + `", "data":{}, "C":` + strconv.Itoa(conf.Conf.C) + `}`
	}

	// if ws.IsServerConn() {
	if err := websocket.Message.Send(ws, msg); err != nil {
		fmt.Println(err)
		// ws.Close()
	}
	// }

	return
}
