package summary

import (
	"strconv"

	"golang.org/x/net/websocket"
)

// 结果反馈
func SendResult(msg string, code int, ws *websocket.Conn) {
	if code == 200 {
		msg = `{"code":200, "message":"success", "data":` + msg + `}`
	} else {
		msg = `{"code":` + strconv.Itoa(code) + `, "message":"` + msg + `", "data":{}}`
	}

	if err := websocket.Message.Send(ws, msg); err != nil {
		panic(err)
	}

	return
}
