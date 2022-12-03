package conf

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/Apipost-Team/runnerGo/tools"
)

var usage = `Usage: runnerGo [Options]

Some Options:
  -r  Rounds of request to run, total requests equal r * n
  -n  Number of simultaneous requests, 0<n<=5000, depends on machine performance
  -j  Url of request har url, use " please
      eg: 
      -j 'https://echo.apipost.cn/json-har.json'
  -t  Timeout for each request in seconds, Default is 10
  -h  This help
  -v  Show verison
`

type Config struct {
	Round     int    // 请求多少轮, 只对Url有效
	N         int    // 并发数
	HarFile   string // 需要请求的har
	UrlNum    int
	TimeOut   int //单次请求超时时间
	StartTime int
	EndTime   int
}

var (
	Conf = Config{}
)

func confError(err error) {
	fmt.Println(usage)
	fmt.Println(err)
	os.Exit(1)
}

func arrangeOptions() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, usage)
	}
	help := flag.Bool("h", false, "")
	round := flag.Int("r", 0, "")
	version := flag.Bool("v", false, "")
	n := flag.Int("n", 0, "")
	timeout := flag.Int("t", 10, "")
	harfile := flag.String("j", "", "")
	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}
	if *version {
		fmt.Println("version is", VERSION)
		os.Exit(0)
	}
	if *n <= 0 {
		confError(errors.New("(-n) Number must be greater than 0."))
	}
	Conf.N = *n
	if *round <= 0 {
		confError(errors.New("(-r) Round must be greater than 0."))
	}
	Conf.Round = *round
	if *timeout <= 0 {
		confError(errors.New("(-t) timeout must be greater than 0."))
	}
	Conf.TimeOut = *timeout

	if *harfile == "" {
		confError(errors.New("(-j) harfile is required."))
	}
	Conf.HarFile = *harfile
	Conf.HarFile = tools.ReplaceQmarks(Conf.HarFile, "")
}

func init() {
	arrangeOptions()
	Conf.UrlNum = Conf.N * Conf.Round
	if Conf.N > 5000 {
		Conf.N = 5000
	}
	if Conf.TimeOut <= 0 || Conf.TimeOut > 60 {
		Conf.TimeOut = 60
	}
	Conf.StartTime = int(tools.GetNowUnixNano())
}
