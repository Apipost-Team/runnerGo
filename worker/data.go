package worker

import runnerHttp "github.com/Apipost-Team/runnerGo/http"

type RawWorkData struct {
	C          int                       `json:"c"`         //并发数
	N          int                       `json:"n"`         //循环次数
	MaxRunTime int                       `json:"t"`         //最大运行时间
	Target_id  string                    `json:"target_id"` //唯一标识
	Data       runnerHttp.HarRequestType //请求数据
}
