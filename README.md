

RunnerGo is a develop tool similar to apache bench (ab).

## Usage
* runnerGo -h

```
Options:
  -r  Rounds of request to run, total requests equal r * n
  -n  Number of simultaneous requests, 0<n<=5000, depends on machine performance
  -j  Specify the har file path or URL for request, use " please
      eg: 
      -j 'https://echo.apipost.cn/json-har.json'
  -t  Specify the time (in milliseconds) to wait for requests to return a response, Default is 10
  -h  Show command line help, including a list of options, and sample use cases.
  -v  Displays the current RunnerGo version
```

* Note: -j Specify the har file path or URL for request. See examples/file-har.json / examples/json-har.json

## Super simple to use
RunnerGo is designed to be the simplest way possible to make stress test. 

```
./runnerGo -n 100 -r 2 -j 'https://echo.apipost.cn/json-har.json'
```

## Examples
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