package worker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/Apipost-Team/runnerGo/conf"
	runnerHttp "github.com/Apipost-Team/runnerGo/http"
	"github.com/Apipost-Team/runnerGo/summary"
)

var (
	rwg       sync.WaitGroup
	config    = conf.Conf
	urlChanel = make(chan [1]runnerHttp.HarRequestType, 30000)
)

func worker() {
	for { // for 循环逐个执行 URL
		data, ok := <-urlChanel
		if !ok {
			return
		}
		summary.ResChanel <- runnerHttp.Do(data[0])
	}
}

func addTask() {
	// 根据URL获取资源
	res, err := http.Get(config.HarFile)
	if err != nil {
		fmt.Println(`{"code":"500", "message":"指定 URL 无法访问"}`)
		os.Exit(1)
	}

	// 读取资源数据 body: []byte
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(`{"code":"501", "message":"指定 URL 的 JSON 数据格式不符合规范"}`)
		os.Exit(1)
	}

	// 解析 har 结构
	var harStruct runnerHttp.HarRequestType
	json.Unmarshal([]byte(string(body)), &harStruct)

	for i := 0; i < config.UrlNum; i++ {
		urlChanel <- [1]runnerHttp.HarRequestType{harStruct}
	}

	for {
		time.Sleep(time.Duration(50) * time.Millisecond)
		if len(urlChanel) == 0 {
			close(urlChanel)
			return
		}
	}
}

func StartWork() {
	rwg.Add(1)
	go addTask()
	go func() {
		summary.HandleRes()
		rwg.Done()
	}()
	for index := 0; index < config.N; index++ {
		go func() {
			worker()
		}()
	}
	rwg.Wait()
}
