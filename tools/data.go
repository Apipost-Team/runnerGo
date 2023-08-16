package tools

type ControlData struct {
	C            int     //并发数
	N            int     //循环次数
	Total        int     //总计发送次数
	TimeOut      int     //单次请求超时时间
	Target_id    string  //唯一标识
	StartTime    int     //开始运行时间
	EndTime      int     //结束运行时间
	MaxRunTime   int     //最大运行时间
	IsCancel     bool    //是否主动取消
	IsRunning    bool    //是否还在运行
	WorkCnt      int32   //运行的进程数量
	WorkTagetCnt int32   //期望达到进程数
	Cnt          int     //实际完成数量
	FailedCnt    int     //失败的数量
	Size         int     //接收数据量
	CostTime     float64 //实际请求花费时间
	ReportTime   int     //多久汇报一次执行进度,单位毫秒
	LogType      int     //日志类型 0 关闭， 1开启全量 2仅错误日志
	LogFilename  string  //错误日志文件路径
	TestDataPath string  //测试数据路径
}

type ReportControlData struct {
	ControlData
	Process float64 //进度
	Qps     float64 //执行次数
	Speed   float64 //下载速度
}
