package summary

import (
	"context"
	"math"
	"sort"
	"strconv"
	"sync"

	// "github.com/Apipost-Team/runnerGo/http"

	"github.com/Apipost-Team/runnerGo/conf"
	"github.com/Apipost-Team/runnerGo/tools"
)

var (
	AnalysisData sync.Map
	ResChanel    = make(chan Res)
)

type Res struct {
	Size         int
	TimeStamp    int
	TotalUseTime float64
	Code         int
	ConnTime     float64
	DNSTime      float64
	ReqTime      float64
	DelayTime    float64
	ResTime      float64
}

type SummaryData struct {
	Target_id             string
	CompleteRequests      int
	FailedRequests        int
	SuccessRequests       int
	TimeToken             float64
	TotalDataSize         int
	AvgDataSize           int
	RequestsPerSec        float64
	SuccessRequestsPerSec float64

	MinUseTime        float64
	MaxUseTime        float64
	AvgUseTime        float64
	CodeDetail        map[string]int
	WaitingTimeDetail map[string]int

	AvgConn  float64
	MaxConn  float64
	MinConn  float64
	AvgDNS   float64
	MaxDNS   float64
	MinDNS   float64
	AvgReq   float64
	MaxReq   float64
	MinReq   float64
	AvgDelay float64
	MaxDelay float64
	MinDelay float64
	AvgRes   float64
	MaxRes   float64
	MinRes   float64
}

func HandleRes(control tools.ControlData, ctx context.Context) SummaryData {
	var (
		// RunOverSignal = make(chan int, 1)
		codeDetail  = make(map[int]int)
		summaryData = SummaryData{
			CodeDetail:        make(map[string]int),
			WaitingTimeDetail: make(map[string]int),
			MinConn:           float64(conf.Conf.TimeOut),
			MinDNS:            float64(conf.Conf.TimeOut),
			MinDelay:          float64(conf.Conf.TimeOut),
			MinReq:            float64(conf.Conf.TimeOut),
			MinUseTime:        float64(conf.Conf.TimeOut),
			MinRes:            float64(conf.Conf.TimeOut),
			Target_id:         control.Target_id,
		}

		waitTimes = make([]float64, 0, control.Total)
	)

OutLable:
	for {
		select {
		case <-ctx.Done():
			break OutLable
		default:

		}
		res, ok := <-ResChanel
		if !ok {
			break
		}
		summaryData.CompleteRequests++
		summaryData.TotalDataSize += res.Size

		// fmt.Println(summaryData.CompleteRequests, "-", conf.Conf.UrlNum, "-")
		if summaryData.CompleteRequests == control.Total {
			close(ResChanel)
		}
		code := res.Code
		if _, ok := codeDetail[code]; ok {
			codeDetail[code]++
		} else {
			codeDetail[code] = 1
		}
		if conf.Conf.EndTime < res.TimeStamp {
			conf.Conf.EndTime = res.TimeStamp
		}
		if control.EndTime < res.TimeStamp {
			control.EndTime = res.TimeStamp
		}
		if code > 299 || code < 200 {
			summaryData.FailedRequests++
		} else {
			summaryData.SuccessRequests++
		}
		summaryData.AvgUseTime += res.TotalUseTime
		summaryData.AvgConn += res.ConnTime
		summaryData.AvgDNS += res.DNSTime
		summaryData.AvgDelay += res.DelayTime
		summaryData.AvgReq += res.ReqTime
		summaryData.AvgRes += res.ResTime

		summaryData.MinUseTime = math.Min(res.TotalUseTime, summaryData.MinUseTime)
		summaryData.MinConn = math.Min(res.ConnTime, summaryData.MinConn)
		summaryData.MinDNS = math.Min(res.DNSTime, summaryData.MinDNS)
		summaryData.MinDelay = math.Min(res.DelayTime, summaryData.MinDelay)
		summaryData.MinReq = math.Min(res.ReqTime, summaryData.MinReq)
		summaryData.MinRes = math.Min(res.ResTime, summaryData.MinRes)

		summaryData.MaxUseTime = math.Max(res.TotalUseTime, summaryData.MaxUseTime)
		summaryData.MaxConn = math.Max(res.ConnTime, summaryData.MaxConn)
		summaryData.MaxDNS = math.Max(res.DNSTime, summaryData.MaxDNS)
		summaryData.MaxDelay = math.Max(res.DelayTime, summaryData.MaxDelay)
		summaryData.MaxReq = math.Max(res.ReqTime, summaryData.MaxReq)
		summaryData.MaxRes = math.Max(res.ResTime, summaryData.MaxRes)
		waitTimes = append(waitTimes, res.TotalUseTime)

	}

	summaryData.AvgUseTime = tools.Decimal2(summaryData.AvgUseTime / float64(control.Total))
	summaryData.AvgConn = tools.Decimal2(summaryData.AvgConn / float64(control.Total))
	summaryData.AvgDNS = tools.Decimal2(summaryData.AvgDNS / float64(control.Total))
	summaryData.AvgDelay = tools.Decimal2(summaryData.AvgDelay / float64(control.Total))
	summaryData.AvgReq = tools.Decimal2(summaryData.AvgReq / float64(control.Total))
	summaryData.AvgRes = tools.Decimal2(summaryData.AvgRes / float64(control.Total))
	summaryData.AvgDataSize = summaryData.TotalDataSize / control.Total

	for k, v := range codeDetail {
		summaryData.CodeDetail[strconv.Itoa(k)] = v
	}

	t := (float64(control.EndTime-control.StartTime) / 10e8)
	summaryData.TimeToken = t
	summaryData.RequestsPerSec = float64(control.Total) / t
	summaryData.SuccessRequestsPerSec = float64(summaryData.SuccessRequests) / t
	sort.Float64s(waitTimes)
	waitTimesL := float64(len(waitTimes))
	tps := []float64{0.1, 0.25, 0.5, 0.75, 0.9, 0.95, 0.99, 0.999, 0.9999}
	tpsL := len(tps)
	for i := 0; i < tpsL; i++ {
		summaryData.WaitingTimeDetail[tools.FloatToPercent(tps[i])] = int(waitTimes[int(waitTimesL*tps[i]-1)])
	}

	// Print(summaryData)
	return summaryData
}
