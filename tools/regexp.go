package tools

import (
	"regexp"
	"strings"
)

var (
	keyValueRegexp      = regexp.MustCompile(`^([\w-]+):\s*(.+)`)
	replaceSpaceRegexp  = regexp.MustCompile(`\s+`)
	ignoreSpaceRegexp   = regexp.MustCompile(`(:|,)\s+`)
	replaceQmarksRegexp = regexp.MustCompile(`"|'`)
)

type ReqData struct {
	Url  string
	Body string
}

func KeyValueRexpGetKV(inputsStr string) []string {
	return keyValueRegexp.FindStringSubmatch(inputsStr)
}

func ParseStr(str string) map[string]string {
	rdataMap := map[string]string{}
	str = string(replaceSpaceRegexp.ReplaceAll([]byte(str), []byte(" ")))
	str = string(ignoreSpaceRegexp.ReplaceAll([]byte(str), []byte("${1}_LL_")))
	str = strings.TrimSpace(str)
	sl := strings.Split(str, " ")
	l := len(sl)
	for index := 0; index < l; index += 2 {
		rdataMap[sl[index]] = strings.Replace(sl[index+1], "_LL_", " ", -1)
	}
	return rdataMap
}

func GetReqData(str string) ReqData {
	reqData := ReqData{}
	if str == "" {
		return reqData
	}
	dataMap := ParseStr(str)
	for k, v := range dataMap {
		if k == "-u" {
			reqData.Url = v
		} else if k == "-d" {
			reqData.Body = v
		}
	}
	return reqData
}

// 替换所有的引号
func ReplaceQmarks(str string, new string) string {
	return string(replaceSpaceRegexp.ReplaceAll([]byte(str), []byte("")))
}
