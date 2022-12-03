package tools

import "time"

func GetNowUnixNano() int64 {
	return time.Now().UnixNano()
}
func GetNowUnix() int64 {
	return time.Now().Unix()
}
func Now() time.Time {
	return time.Now()
}
