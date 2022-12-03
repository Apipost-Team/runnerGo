package logger

import (
	"flag"
	"log"
	"os"
)

var (
	Log *log.Logger
)

func init() {
	// set location of log file
	var logpath = "./info.log"

	flag.Parse()
	var file, err1 = os.Create(logpath)

	if err1 != nil {
		panic(err1)
	}
	// Log = log.New(file, "", log.LstdFlags|log.Lshortfile)
	Log = log.New(file, "", 0)
}
