

runnerGo 是一个可以帮助你获取http服务器压力测试性能指标的工具，有点像Apache的ab，不同的是，它可以帮你发送携带不同参数的请求，这样你就可以便捷地重放线上的真实请求。

## 参数说明
* runnerGo -h

```
Options:
  -r  Rounds of request to run, total requests equal r * n
  -n  Number of simultaneous requests, 0<n<=900, depends on machine performance
  -j  Url of request har url, use " please
      eg: 
      -j 'https://echo.apipost.cn/json-har.json'
  -t  Timeout for each request in seconds, Default is 10
  -h  This help
  -v  Show verison
```

* 注意: -j 是实现发送的 har json路径，文件详细内容，可参照examples/file-har.json 和 examples/json-har.json

## 一些例子
* 1: runnerGo -n 100 -r 2 -j https://echo.apipost.cn/json-har.json

## 结果展示
```json
{
    "Summary": {
        "CompleteRequests": 8,
        "FailedRequests": 0,
        "TimeToken": 0.760136,
        "TotalDataSize": 3728,
        "AvgDataSize": 466,
        "MaxUseTime": 312,
        "MinUseTime": 10,
        "AvgUseTime": 233,
        "RequestsPerSec": 10.524432469979056,
        "SuccessRequestsPerSec": 10.524432469979056
    },
    "WaitingTimeDetail": {
        "10.00%": 105,
        "25.00%": 122,
        "50.00%": 239,
        "75.00%": 295,
        "90.00%": 301,
        "95.00%": 301,
        "99.00%": 301,
        "99.90%": 301,
        "99.99%": 301
    },
    "CodeDetail": {
        "200": 8
    },
    "Times": {
        "dns": {
            "MinDNS": 0,
            "AvgDNS": 1.375,
            "MaxDNS": 3
        },
        "conn": {
            "MinConn": 0,
            "AvgConn": 96.5,
            "MaxConn": 139
        },
        "wait": {
            "MinDelay": 10,
            "AvgDelay": 136.25,
            "MaxDelay": 173
        },
        "resp": {
            "MinRes": 0,
            "AvgRes": 0,
            "MaxRes": 0
        }
    }
}
```