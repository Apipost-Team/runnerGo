package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	//_ "net/http/pprof"

	"github.com/Apipost-Team/runnerGo/request"
	"github.com/gorilla/websocket"

	runnerHttp "github.com/Apipost-Team/runnerGo/http"
)

var WebsocketCnt int32 = 0 //连接数量
func delayExit(delay time.Duration) {
	time.Sleep(delay * time.Second)
	if WebsocketCnt < 1 {
		log.Println("free too long, exit")
		os.Exit(0)
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("发生错误:", err)
		}
	}()

	var serverPort int
	var isAutoExit int

	flag.IntVar(&serverPort, "p", 10397, "server port， default：10397")
	flag.IntVar(&isAutoExit, "a", 0, "is auto exit， default：0")
	flag.Parse()
	fmt.Printf("server port %d and is auto exit %d", serverPort, isAutoExit)

	if isAutoExit > 0 {
		go delayExit(30) //30s不使用退出
	}

	http.HandleFunc("/quit", func(w http.ResponseWriter, r *http.Request) {
		//增加退出代码
		log.Println("user quit")
		os.Exit(0)
	})
	http.HandleFunc("/send", func(w http.ResponseWriter, r *http.Request) {
		//增加代理发送功能
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(string(body))
		var p runnerHttp.HarRequestType
		err = json.Unmarshal(body, &p)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(p)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{}"))
	})
	http.HandleFunc("/websocket", func(w http.ResponseWriter, r *http.Request) {
		// 升级HTTP连接为WebSocket连接
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		defer conn.Close()

		//记录客户端数量
		atomic.AddInt32(&WebsocketCnt, 1)
		defer func() {
			atomic.AddInt32(&WebsocketCnt, -1) //连接数减少
		}()

		fmt.Println("conn:", WebsocketCnt)
		defer func() {
			fmt.Println("close:", WebsocketCnt)
		}()

		var sendChan = make(chan string)

		go request.ReadAndDo(sendChan, conn) //读取并执行命令

		for {
			msg, ok := <-sendChan
			if !ok {
				break
			}
			if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
				fmt.Println("write")
				fmt.Println(err)
				break
			}
		}

		fmt.Println("websocket is closed", WebsocketCnt)
		if isAutoExit > 0 {
			//断开，3s后重启
			go delayExit(3) //30s不使用退出
		}
	})

	if err := http.ListenAndServe(":"+strconv.Itoa(serverPort), nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
