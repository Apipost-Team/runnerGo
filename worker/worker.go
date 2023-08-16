package worker

import (
	"context"
	"crypto/tls"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"sync/atomic"
	"time"

	runnerHttp "github.com/Apipost-Team/runnerGo/http"
	"github.com/Apipost-Team/runnerGo/summary"
	"github.com/Apipost-Team/runnerGo/tools"
)

type NullWriter int

func (NullWriter) Write([]byte) (int, error) { return 0, nil }

// 用默认浏览器打开指定URL
func OpenUrl(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"

	}

	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

func Process(control *tools.ControlData, data runnerHttp.HarRequestType, sendChan chan<- string) {
	//control  初始化
	control.StartTime = int(tools.GetNowUnixNano())
	control.EndTime = control.StartTime
	control.IsRunning = true //设置为启动
	control.IsCancel = false //设置为取消
	control.WorkCnt = 0      //设置工作进程为0

	var urlChanel = make(chan runnerHttp.HarRequestType, 8)    //8个url任务列表,带缓冲
	var resultChanel = make(chan summary.Res, 16)              //返回结果列表，带缓冲
	var rLog *log.Logger                                       //日志文件
	ctx, cancelAll := context.WithCancel(context.Background()) //控制其他进程退出

	//设置并发任务消费,需要连接池
	defaultCipherSuites := []uint16{0xc02f, 0xc030, 0xc02b, 0xc02c, 0xcca8, 0xcca9, 0xc013, 0xc009,
		0xc014, 0xc00a, 0x009c, 0x009d, 0x002f, 0x0035, 0xc012, 0x000a}
	tr := &http.Transport{
		MaxConnsPerHost:    0,
		IdleConnTimeout:    20 * time.Second,
		DisableKeepAlives:  false,
		DisableCompression: false,
		TLSClientConfig: &tls.Config{
			CipherSuites:       append(defaultCipherSuites[8:], defaultCipherSuites[:8]...),
			InsecureSkipVerify: true,
		}, //绕过特征检测
		TLSHandshakeTimeout: 10 * time.Second,
	}

	//程序错误处理
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("执行任务出错,忽略退出", r)
			msg := fmt.Sprintf("\"code\":501, \"message\": \"%q\", \"data\":{\"Target_id\":\"%s\"}}", r, control.Target_id)
			sendChan <- msg
		}
	}()

	//工作接收清理工作
	defer func() {
		control.IsRunning = false
		cancelAll() //取消所有任务
		fmt.Println("action main quit", *control)
	}() //设置为执行完成

	//判断是否要创建文本日志
	if control.LogType > 0 {
		f, err := os.CreateTemp("", "runnergo_")
		if err != nil {
			//出错了，无法开启日志
			rLog.SetFlags(0)                //屏蔽所有日志
			rLog.SetOutput(new(NullWriter)) //设置空接口
		} else {
			defer f.Close() //关闭日志
			rLog = log.New(f, "", log.Ltime)
			control.LogFilename = f.Name() //错误日志文件
			fmt.Printf("log file:%s", f.Name())
		}
	}

	//运行是控制，取消，超时，控制进程数量，定时汇报
	go doControl(control, ctx, sendChan, tr, urlChanel, resultChanel, rLog)

	//启动工作进程
	for i := 0; i < control.C; i++ {
		go doWork(control, ctx, tr, i, urlChanel, resultChanel, rLog)
	}

	//创建工作任务
	go func(urlChanel chan<- runnerHttp.HarRequestType, data runnerHttp.HarRequestType, control *tools.ControlData, ctx context.Context) {
		defer func() {
			close(urlChanel) //关闭请求产生
			if r := recover(); r != nil {
				fmt.Println("error add task", r)
			}
		}()

		has_var := false
		var test_data_list []map[string]string
		if len(control.TestDataPath) > 0 {
			//读取测试数据
			test_data_list = readCsv(control.TestDataPath)
		}
		test_data_len := len(test_data_list)       //测试数据长度
		empty_test_data := make(map[string]string) //空数据

		//{{变量}} 检查
		json_str, err := json.Marshal(data)
		if err != nil {
			control.IsCancel = true //设置为取消
			fmt.Println("action add task error", err)
			return
		}
		raw_data_str := string(json_str)
		pattern := `{{.*?}}`
		match, _ := regexp.MatchString(pattern, raw_data_str)
		if match {
			has_var = true
		}
		fmt.Println("has_var", has_var, "test_data_len", test_data_len)

		is_forever := false
		if control.Total <= 0 && control.MaxRunTime > 0 {
			is_forever = true //永久执行
		}

		//按次数循环模式

		for i := 0; is_forever || i < control.Total; i++ {
			if control.IsCancel || (!control.IsRunning) {
				fmt.Println("action add task quit")
				break
			}

			//变量替换逻辑
			if has_var {
				//替换变量
				var new_data_str string
				if test_data_len < 1 {
					new_data_str = replaceVariables(raw_data_str, empty_test_data)
				} else {
					new_data_str = replaceVariables(raw_data_str, test_data_list[i%test_data_len])
				}

				err = json.Unmarshal([]byte(new_data_str), &data) //替换变量
				if err != nil {
					control.IsCancel = true //设置为取消
					fmt.Println("data error", err)
					continue
				}
			}

			data.Seq = i //设置请求序列

			select {
			case urlChanel <- data:
				//log.Printf("send data %d", i)
			case <-ctx.Done():
				//防止写入死锁
				fmt.Println("action reciver exit")
				return
			}
		}

	}(urlChanel, data, control, ctx)

	//统计结果
	res := summary.HandleRes(control, resultChanel)
	jsonRes, err := json.Marshal(res)

	var msg string
	if err != nil {
		msg = `{"code":501, "message":"` + string(err.Error()) + `", "data":{"Target_id":"` + control.Target_id + `"}}`
	} else {
		msg = `{"code":200, "message":"success", "data":` + string(jsonRes) + `}`
	}

	sendChan <- msg
}

func doWork(control *tools.ControlData, ctx context.Context, tr *http.Transport, workId int, urlChanel <-chan runnerHttp.HarRequestType, resultChanel chan<- summary.Res, rLog *log.Logger) {
	defer func() {
		if control.WorkCnt < 1 {
			//最后一个工作进程，清理管道
			fmt.Printf("close channel by work %d and workcnt %d", workId, control.WorkCnt)
			close(resultChanel)
		}
		if r := recover(); r != nil {
			fmt.Println("work error", workId, r)
		}
	}()

	//设置数量
	atomic.AddInt32(&(control.WorkCnt), 1) //进程数加1
	defer func() {
		atomic.AddInt32(&(control.WorkCnt), -1) //工作线程减1
	}()

	//初始化 httpclient
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Duration(control.TimeOut) * time.Second,
	}

	for { // for 循环逐个执行 URL
		if control.IsCancel || (!control.IsRunning) {
			break
		} else {
			data, ok := <-urlChanel
			if !ok {
				//工作已经执行完成
				break
			}

			if data.Seq < 0 {
				//只有控制信息seq会为负数
				break
			}

			result, err := runnerHttp.Do(client, data)
			//日志记录
			if control.LogType > 0 {
				if err != nil {
					rLog.Printf("seq:%d\twork:%d\tcode:%d\tcost:%.1f\tmsg:%s", data.Seq, workId, result.Code, result.TotalUseTime, err.Error())
				} else if control.LogType < 2 {
					rLog.Printf("seq:%d\twork:%d\tcode:%d\tcost:%.1f\tmsg:%s", data.Seq, workId, result.Code, result.TotalUseTime, "ok")
				}
			}

			//写入时循环检查
			select {
			case resultChanel <- result:
				break
			case <-ctx.Done():
				break
			}
		}
	}
}

func doControl(control *tools.ControlData, ctx context.Context, sendChan chan<- string, tr *http.Transport, urlChanel chan runnerHttp.HarRequestType, resultChanel chan<- summary.Res, rLog *log.Logger) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("error control", r)
		}
	}()

	timeChan := time.After(time.Second * time.Duration(control.MaxRunTime)) //超时设置
	checkCnt := 0                                                           //检查次数
	checkInterval := 50                                                     //检查间隔
	reportInterval := 0                                                     //反馈间隔
	if control.ReportTime > 0 {
		reportInterval = int(float64(control.ReportTime)/float64(checkInterval) + 0.5) //向上取整数
		if reportInterval < 1 {
			reportInterval = 1 //report时间最小50毫秒
		}
	}

	fmt.Println("run timeout", control.MaxRunTime, "reportInterval", reportInterval, "reportTime", control.ReportTime)

	oldControl := *control                      //旧的控制信息
	var process, qps, speed, spend_time float64 //进度,执行次数,下载速度

OutCancel:
	for {
		select {
		case <-timeChan:
			control.IsCancel = true //设置为取消
			fmt.Println("action timout")
			break OutCancel
		case <-time.After(time.Duration(checkInterval) * time.Millisecond):
			checkCnt++ //检查次数+1
			//确定退出
			if control.IsCancel || (!control.IsRunning) { //取消或者支持完成直接退出
				if control.IsCancel {
					fmt.Println("action cancel")
				} else {
					fmt.Println("action end")
				}
				break OutCancel
			}

			//检查report
			if reportInterval > 0 && (checkCnt%reportInterval == 0) {

				if control.Total < 1 {
					//按时间统计进度
					process = float64(control.EndTime-control.StartTime) / 10e8 / float64(control.MaxRunTime)
				} else {
					process = float64(control.Cnt) / float64(control.Total)
				}

				spend_time = (float64(control.EndTime) - float64(oldControl.EndTime)) / 10e8
				if spend_time > 0 {
					qps = float64(control.Cnt-oldControl.Cnt) / spend_time
					speed = float64(control.Size-oldControl.Size) / 1024 / 1024 / spend_time
				} else {
					spend_time = float64(control.EndTime-control.StartTime) / 10e8
					if spend_time > 0 {
						qps = float64(control.Cnt) / spend_time
						speed = float64(control.Size) / 1024 / 1024 / spend_time
					}
				}

				reportData := tools.ReportControlData{
					*control, process, qps, speed,
				}

				jsonRes, err := json.Marshal(reportData)
				if err != nil {
					fmt.Println(err.Error())
					continue
				}

				msg := `{"code":202, "message":"success", "data":` + string(jsonRes) + `}`
				sendChan <- msg //发送统计信息
			}

			//检查是否需要增加进程
			if control.WorkTagetCnt > 0 {
				diffCnt := int(control.WorkTagetCnt - control.WorkCnt) //目标和实际差距
				if diffCnt > 0 {
					//增加进程
					for i := 0; i < diffCnt; i++ {
						go doWork(control, ctx, tr, i+int(control.WorkTagetCnt), urlChanel, resultChanel, rLog)
					}
					control.WorkTagetCnt = 0 //启动完成，标记
				} else if diffCnt < 0 {
					//减少进程
					diffCnt = -diffCnt

					for i := 0; i < diffCnt; i++ {
						emptyData := runnerHttp.HarRequestType{}
						emptyData.Seq = -1
						urlChanel <- emptyData
					}
					control.WorkTagetCnt = 0 //可能阻塞,标记
				}
			}

		}
	}
}

func Request(p runnerHttp.HarRequestType) {
	//初始化 httpclient
	// client := &http.Client{
	// 	Timeout: time.Duration(10) * time.Second,
	// }

	//result, err := runnerHttp.Do(client, p)
}

func replaceVariables(input string, data map[string]string) string {
	// 正则表达式，匹配 {{variable}} 格式的变量,然后通过map中数据替换
	re := regexp.MustCompile(`{{(.*?)}}`)
	output := re.ReplaceAllStringFunc(input, func(match string) string {
		key := match[2 : len(match)-2] // 移除{{}}包围的变量名
		value, exists := data[key]
		if exists {
			return value
		}
		return "" // 如果变量名不存在于map中，则替换为空字符串
	})
	return output
}

func readCsv(path string) []map[string]string {
	//读取csv文件
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Failed to open CSV file:", err)
		return nil
	}
	defer file.Close()

	// 创建 CSV 读取器
	reader := csv.NewReader(file)

	// 读取 CSV 数据
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Failed to read CSV data:", err)
		return nil
	}

	// 将 CSV 数据转换为 []map[string]string
	var data []map[string]string
	headers := records[0]
	for _, record := range records[1:] {
		row := make(map[string]string)
		for i, value := range record {
			row[headers[i]] = value
		}
		data = append(data, row)
	}

	return data
}
