package conf

import (
	"flag"
	"fmt"
	"os"
)

var usage = `Usage: runnerGo [Options]

Some Options:
	-n 	requests     Number of requests to perform
	-c 	concurrency  Number of multiple requests to make at a time
	-t  Timeout for each request in seconds, Default is 10
	-h  This help
	-v  Show verison
`

type Config struct {
	C         int
	UrlNum    int
	TimeOut   int //单次请求超时时间
	StartTime int
	EndTime   int
}

var (
	Conf = Config{}
)

func arrangeOptions() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, usage)
	}
	help := flag.Bool("h", false, "")
	version := flag.Bool("v", false, "")
	timeout := flag.Int("t", 60, "")
	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}
	if *version {
		fmt.Println("version is", VERSION)
		os.Exit(0)
	}

	Conf.TimeOut = *timeout
}

func init() {
	arrangeOptions()
	if Conf.TimeOut <= 0 || Conf.TimeOut > 60 {
		Conf.TimeOut = 60
	}
}
