package tools

type ControlData struct {
	C          int    //并发数
	N          int    //循环次数
	Total      int    //总计发送次数
	TimeOut    int    //单次请求超时时间
	Target_id  string //唯一标识
	StartTime  int    //开始运行时间
	EndTime    int    //结束运行时间
	MaxRunTime int    //最大运行时间
	IsCancel   bool   //是否主动取消
	IsRunning  bool   //是否还在运行
	WorkCnt    int32  //运行的进程数量
}
