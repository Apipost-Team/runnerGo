package summary

import (
	"encoding/json"
	"fmt"
	"os"
	"text/template"
)

var (
	htmlTemplate = `{
		"Summary":{
			"CompleteRequests":{{ .CompleteRequests }},
			"FailedRequests":{{ .FailedRequests }},
			"TimeToken":{{ .TimeToken }},
			"TotalDataSize":{{ .TotalDataSize }},
			"AvgDataSize":{{ .AvgDataSize }},
			"MaxUseTime":{{ .MaxUseTime }},
			"MinUseTime":{{ .MinUseTime }},
			"AvgUseTime":{{ .AvgUseTime }},
			"RequestsPerSec":{{ .RequestsPerSec }},
			"SuccessRequestsPerSec":{{ .SuccessRequestsPerSec }}
		},
		"WaitingTimeDetail":{{ formatMap .WaitingTimeDetail }},
		"CodeDetail":{{ formatMap .CodeDetail }},
		"Times":{
			"dns":{
				"MinDNS":{{.MinDNS}},
				"AvgDNS":{{.AvgDNS}},
				"MaxDNS":{{.MaxDNS}}
			},
			"conn":{
				"MinConn":{{.MinConn}},
				"AvgConn":{{.AvgConn}},
				"MaxConn":{{.MaxConn}}
			},
			"wait":{
				"MinDelay":{{.MinDelay}},
				"AvgDelay":{{.AvgDelay}},
				"MaxDelay":{{.MaxDelay}}
			},
			"resp":{
				"MinRes":{{.MinRes}},
				"AvgRes":{{.AvgRes}},
				"MaxRes":{{.MaxRes}}
			}
		}
	}`
)

func formatMap(data map[string]int) string {
	dataType, _ := json.Marshal(data)
	dataString := string(dataType)
	return dataString
}

var tmplFuncMap = template.FuncMap{
	"formatMap": formatMap,
}

func Print(summaryData SummaryData) {

	tmpl, err := template.New("test").Funcs(tmplFuncMap).Parse(htmlTemplate)
	if err != nil {
		fmt.Println(`{"code":"510", "message":"操作失败,稍后再试(510)"}`)
		os.Exit(1)
	}
	err = tmpl.Execute(os.Stdout, summaryData)
	if err != nil {
		fmt.Println(`{"code":"511", "message":"操作失败,稍后再试(511)"}`)
		os.Exit(1)
	}

}
