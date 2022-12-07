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
}

// var (
// 	htmlTemplate = `{
// 		"Summary":{
// 			"CompleteRequests":{{ .CompleteRequests }},
// 			"FailedRequests":{{ .FailedRequests }},
// 			"TimeToken":{{ .TimeToken }},
// 			"TotalDataSize":{{ .TotalDataSize }},
// 			"AvgDataSize":{{ .AvgDataSize }},
// 			"MaxUseTime":{{ .MaxUseTime }},
// 			"MinUseTime":{{ .MinUseTime }},
// 			"AvgUseTime":{{ .AvgUseTime }},
// 			"RequestsPerSec":{{ .RequestsPerSec }},
// 			"SuccessRequestsPerSec":{{ .SuccessRequestsPerSec }}
// 		},
// 		"WaitingTimeDetail":{{ formatMap .WaitingTimeDetail }},
// 		"CodeDetail":{{ formatMap .CodeDetail }},
// 		"Times":{
// 			"dns":{
// 				"MinDNS":{{.MinDNS}},
// 				"AvgDNS":{{.AvgDNS}},
// 				"MaxDNS":{{.MaxDNS}}
// 			},
// 			"conn":{
// 				"MinConn":{{.MinConn}},
// 				"AvgConn":{{.AvgConn}},
// 				"MaxConn":{{.MaxConn}}
// 			},
// 			"wait":{
// 				"MinDelay":{{.MinDelay}},
// 				"AvgDelay":{{.AvgDelay}},
// 				"MaxDelay":{{.MaxDelay}}
// 			},
// 			"resp":{
// 				"MinRes":{{.MinRes}},
// 				"AvgRes":{{.AvgRes}},
// 				"MaxRes":{{.MaxRes}}
// 			}
// 		}
// 	}`
// )

// func formatMap(data map[string]int) string {
// 	dataType, _ := json.Marshal(data)
// 	dataString := string(dataType)
// 	return dataString
// }

// var tmplFuncMap = template.FuncMap{
// 	"formatMap": formatMap,
// }

// func Print(summaryData SummaryData) {

// 	tmpl, err := template.New("test").Funcs(tmplFuncMap).Parse(htmlTemplate)
// 	if err != nil {
// 		ErrorPrint(`{"code":"510", "message":"操作失败,稍后再试(510)"}`)
// 	}
// 	err = tmpl.Execute(os.Stdout, summaryData)
// 	if err != nil {
// 		ErrorPrint(`{"code":"511", "message":"操作失败,稍后再试(511)"}`)
// 	}

// }

// func ErrorPrint(err string) {
// 	// fmt.Fprintln(err)
// 	fmt.Fprintln(os.Stdout, err)
// }
