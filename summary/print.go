package summary

import (
	"encoding/json"
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

// 	htmlTemplate = `
// Summary:
//   Complete requests:		{{ .CompleteRequests }}
//   Failed requests:		{{ .FailedRequests }}
//   Time taken (s):		{{ .TimeToken }}
//   Total data size (Byte):	{{ .TotalDataSize }}
//   Data size/request (Byte):	{{ .AvgDataSize }}
//   Max use time (ms):		{{.MaxUseTime}}
//   Min use time (ms):		{{.MinUseTime}}
//   Average use time (ms):	{{.AvgUseTime}}
//   Requests/sec:			{{ .RequestsPerSec }}
//   SuccessRequests/sec:		{{ .SuccessRequestsPerSec }}

// Percentage of waiting time (ms):
// {{ formatMap .WaitingTimeDetail }}

// Time detail (ms)
//   item		min		mean		max
//   dns		{{.MinDNS}}		{{.AvgDNS}}		{{.MaxDNS}}
//   conn		{{.MinConn}}		{{.AvgConn}}		{{.MaxConn}}
//   wait		{{.MinDelay}}		{{.AvgDelay}}		{{.MaxDelay}}
//   resp		{{.MinRes}}		{{.AvgRes}}		{{.MaxRes}}

// Response Time histogram (code: requests):
// {{ formatMap .CodeDetail }}
// `
)

func formatMap(data map[string]int) string {
	dataType, _ := json.Marshal(data)
	dataString := string(dataType)
	return dataString
	// var keys []string

	// for k, _ := range data {
	// 	keys = append(keys, k)

	// }
	// sort.Strings(keys)

	// res := new(bytes.Buffer)
	// for _, k := range keys {
	// 	res.WriteString(fmt.Sprintf("%s:\t\t%d\n", k, data[k]))
	// 	// res.WriteString(fmt.Sprintf("  %s:\t\t%d\n", k, data[k]))
	// }
	// // fmt.Printf("%+v\n", `{"ids":123}`)
	// // fmt.Printf("%+v\n", res.String())
	// return res.String()
}

var tmplFuncMap = template.FuncMap{
	"formatMap": formatMap,
}

func Print(summaryData SummaryData) {

	tmpl, err := template.New("test").Funcs(tmplFuncMap).Parse(htmlTemplate)
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(os.Stdout, summaryData)
	if err != nil {
		panic(err)
	}

}
