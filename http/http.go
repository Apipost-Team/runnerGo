package http

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/http/httptrace"
	"net/url"
	netulr "net/url"
	"os"
	"strings"
	"time"

	"github.com/Apipost-Team/runnerGo/conf"
	"github.com/Apipost-Team/runnerGo/summary"
	"github.com/Apipost-Team/runnerGo/tools"
)

// Part of the HAR JSON Data structure.
// The header of a HTTP request or response is an array of name-value Tuples.
type harHeaderType struct {
	Name  string
	Value string
}

// Part of the HAR JSON Data structure.
// The information stored for one cookie.
type harCookieType struct {
	Name     string
	Value    string
	Path     string
	Domain   string
	Expires  string
	HttpOnly bool
	Secure   bool
}

type postDataParaType struct {
	Name        string
	Value       string
	FileName    string
	Type        string
	ContentType string
	Comment     string
}
type postDataType struct {
	MimeType string
	Params   []postDataParaType
	Text     string
	Comment  string
}

type HarRequestType struct {
	Method      string //[5]string{"GET", "POST", "PUT", "DELETE", "HEAD"}
	Url         string
	Mode        string
	HttpVersion string
	Cookies     []harCookieType
	Headers     []harHeaderType
	QueryString []interface{}
	PostData    postDataType
	HeadersSize int
	BodySize    int
}

type RunnerGoType struct {
	Method  string
	Url     string
	Mode    string
	Cookies []harCookieType
	Headers map[string]string
}

const (
	clientsN int = 2
)

var (
	HttpClients []*http.Client
	config      = conf.Conf
)

func init() {
	for i := 0; i < clientsN; i++ {
		HttpClients = append(HttpClients, creteHttpClient())
	}
}

func creteHttpClient() *http.Client {
	client := &http.Client{
		Transport: &http.Transport{
			MaxConnsPerHost:     config.N/clientsN + 128,
			MaxIdleConnsPerHost: config.N/clientsN + 128,
			DisableKeepAlives:   false,
			DisableCompression:  false,
		},
		Timeout: time.Duration(config.TimeOut) * time.Second,
	}
	return client
}

func Do(harStruct HarRequestType) summary.Res {
	var code int
	var size, tmpt int64
	var dnsStart, connStart, respStart, reqStart, delayStart int64
	var dnsDuration, connDuration, respDuration, reqDuration, delayDuration int64

	// 创建 压测数据对象
	var runnerGoStruct RunnerGoType

	// 设置一些基本项
	runnerGoStruct.Url = harStruct.Url
	runnerGoStruct.Method = harStruct.Method
	runnerGoStruct.Cookies = harStruct.Cookies
	runnerGoStruct.Mode = harStruct.Mode

	// 校验 URL
	if runnerGoStruct.Url == "" || (strings.ToLower(runnerGoStruct.Url)[:7] != "http://" && strings.ToLower(runnerGoStruct.Url)[:8] != "https://") {
		fmt.Println(`{"code":"502", "message":"请输入正常的 URL(` + runnerGoStruct.Url + `)"}`)
		os.Exit(1)
	}

	// 校验 method
	if runnerGoStruct.Url == "" {
		runnerGoStruct.Url = "GET"
	}

	// 校验 mode 并生成header+body
	runnerGoStruct.Headers = make(map[string]string)
	var req *http.Request
	var newReqErr error

	switch runnerGoStruct.Mode {
	case "form-data":
		bodyBuf := &bytes.Buffer{}
		bodyWriter := multipart.NewWriter(bodyBuf)
		isEmptyBody := true
		for _, v := range harStruct.PostData.Params {
			v.Name = strings.TrimSpace(v.Name)

			if v.Name != "" {
				isEmptyBody = false
				if strings.ToLower(strings.TrimSpace(v.Type)) == "file" && strings.TrimSpace(v.Value)[:1] == "@" {
					v.Value = strings.TrimSpace(v.Value)
					filePath := v.Value[1:]

					fileInfo, e := os.Stat(filePath)
					if e != nil {
						fmt.Println(`{"code":"503", "message":"参数指定的文件路径不存在(` + filePath + `)"}`)
						os.Exit(1)
					}
					if fileInfo.IsDir() {
						fmt.Println(`{"code":"504", "message":"参数指定的路径是目录而不是一个文件(` + filePath + `)"}`)
						os.Exit(1)
					}

					fileWriter, e := bodyWriter.CreateFormFile(v.Name, fileInfo.Name())
					if e != nil {
						fmt.Println(`{"code":"505", "message":"临时文件创建失败(` + filePath + `)"}`)
						os.Exit(1)
					}

					fileOpen, e := os.Open(filePath)
					if e != nil {
						fmt.Println(`{"code":"506", "message":"临时文件创建失败(` + filePath + `)"}`)
						os.Exit(1)
					}
					defer fileOpen.Close()
					_, e = io.Copy(fileWriter, fileOpen)
					if e != nil {
						fmt.Println(`{"code":"507", "message":"临时文件创建失败(` + filePath + `)"}`)
						os.Exit(1)
					}
				} else {
					bodyWriter.WriteField(v.Name, v.Value)
				}
			}
		}
		bodyWriter.Close() // 这句话必不可少,且前面不能加 defer

		// 参数不为空的话,设置请求头
		if isEmptyBody != true {
			runnerGoStruct.Headers["content-type"] = bodyWriter.FormDataContentType()
			req, newReqErr = http.NewRequest(runnerGoStruct.Method, runnerGoStruct.Url, bodyBuf)
		} else {
			req, newReqErr = http.NewRequest(runnerGoStruct.Method, runnerGoStruct.Url, nil)
		}
		break
	case "urlencoded":
		runnerGoStruct.Headers["content-type"] = "application/x-www-form-urlencoded"
		bodyBuf := url.Values{}
		for _, v := range harStruct.PostData.Params {
			v.Name = strings.TrimSpace(v.Name)

			if v.Name != "" {
				bodyBuf.Set(v.Name, v.Value)
			}
		}

		req, newReqErr = http.NewRequest(runnerGoStruct.Method, runnerGoStruct.Url, strings.NewReader(bodyBuf.Encode()))
		break
	case "json":
		runnerGoStruct.Headers["content-type"] = "application/json"
		if harStruct.PostData.Text != "" {
			req, newReqErr = http.NewRequest(runnerGoStruct.Method, runnerGoStruct.Url, bytes.NewBuffer([]byte(harStruct.PostData.Text)))
		} else {
			req, newReqErr = http.NewRequest(runnerGoStruct.Method, runnerGoStruct.Url, nil)
		}
		break
	case "xml":
		runnerGoStruct.Headers["content-type"] = "application/xml"
		if harStruct.PostData.Text != "" {
			req, newReqErr = http.NewRequest(runnerGoStruct.Method, runnerGoStruct.Url, bytes.NewBuffer([]byte(harStruct.PostData.Text)))
		} else {
			req, newReqErr = http.NewRequest(runnerGoStruct.Method, runnerGoStruct.Url, nil)
		}
		break
	case "javascript":
		runnerGoStruct.Headers["content-type"] = "application/javascript"
		if harStruct.PostData.Text != "" {
			req, newReqErr = http.NewRequest(runnerGoStruct.Method, runnerGoStruct.Url, bytes.NewBuffer([]byte(harStruct.PostData.Text)))
		} else {
			req, newReqErr = http.NewRequest(runnerGoStruct.Method, runnerGoStruct.Url, nil)
		}
		break
	case "plain":
		runnerGoStruct.Headers["content-type"] = "text/plain"
		if harStruct.PostData.Text != "" {
			req, newReqErr = http.NewRequest(runnerGoStruct.Method, runnerGoStruct.Url, bytes.NewBuffer([]byte(harStruct.PostData.Text)))
		} else {
			req, newReqErr = http.NewRequest(runnerGoStruct.Method, runnerGoStruct.Url, nil)
		}
		break
	case "html":
		runnerGoStruct.Headers["content-type"] = "text/html"
		if harStruct.PostData.Text != "" {
			req, newReqErr = http.NewRequest(runnerGoStruct.Method, runnerGoStruct.Url, bytes.NewBuffer([]byte(harStruct.PostData.Text)))
		} else {
			req, newReqErr = http.NewRequest(runnerGoStruct.Method, runnerGoStruct.Url, nil)
		}
		break
	default:
		req, newReqErr = http.NewRequest(runnerGoStruct.Method, runnerGoStruct.Url, nil)
		break
	}

	if newReqErr != nil {
		fmt.Println(`{"code":"508", "message":"操作失败,稍后再试(508)"}`)
		os.Exit(1)
	}

	// 设置请求头
	for _, v := range harStruct.Headers {
		runnerGoStruct.Headers[v.Name] = v.Value
	}

	for k, v := range runnerGoStruct.Headers {
		if strings.ToLower(k) == "host" {
			req.Host = v
		} else {
			req.Header.Set(k, v)
		}
	}
	// req.Header.Set("User-Agent", browser.Random())

	trace := &httptrace.ClientTrace{
		DNSStart: func(info httptrace.DNSStartInfo) {
			dnsStart = tools.GetNowUnixNano()
		},
		DNSDone: func(dnsInfo httptrace.DNSDoneInfo) {
			dnsDuration = tools.GetNowUnixNano() - dnsStart
		},
		GetConn: func(h string) {
			connStart = tools.GetNowUnixNano()
		},
		GotConn: func(connInfo httptrace.GotConnInfo) {
			tmpt = tools.GetNowUnixNano()
			if !connInfo.Reused {
				connDuration = tmpt - connStart
			}
			reqStart = tmpt
		},
		WroteRequest: func(w httptrace.WroteRequestInfo) {
			tmpt = tools.GetNowUnixNano()
			reqDuration = tmpt - reqStart
			delayStart = tmpt
		},
		GotFirstResponseByte: func() {
			tmpt = tools.GetNowUnixNano()
			delayDuration = tmpt - delayStart
			respStart = tmpt
		},
	}
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
	tStart := tools.GetNowUnixNano()

	client := HttpClients[rand.Intn(clientsN)]
	response, err := client.Do(req)

	tEnd := tools.Now()
	if response != nil {
		if response.ContentLength > -1 {
			size = response.ContentLength
		} else {
			size = 0
		}
		code = response.StatusCode
		bSize := 32 * 1024
		if int64(bSize) > size {
			if size < 1 {
				bSize = 1
			} else {
				bSize = int(size)
			}
		}
		buf := make([]byte, bSize)
		io.CopyBuffer(ioutil.Discard, response.Body, buf)
		response.Body.Close()
	} else {
		code = 503
		if err, ok := err.(*netulr.Error); ok {
			if err.Timeout() {
				code = 504
			}
		}
	}

	respDuration = tEnd.UnixNano() - respStart

	return summary.Res{
		Size:         int(size),
		TimeStamp:    int(tEnd.UnixNano()),
		TotalUseTime: float64((tEnd.UnixNano() - tStart) / 10e5),
		Code:         code,
		ConnTime:     float64(connDuration / 10e5),
		DNSTime:      float64(dnsDuration / 10e5),
		ReqTime:      float64(reqDuration / 10e5),
		DelayTime:    float64(delayDuration / 10e5),
		ResTime:      float64(respDuration / 10e5),
	}

}
