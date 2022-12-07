

RunnerGo is a develop tool similar to apache bench (ab).

## Usage
* runnerGo -h

```
Options:
    -n 	requests     Number of requests to perform
    -c 	concurrency  Number of multiple requests to make at a time
    -t  Timeout for each request in seconds, Default is 60
    -h  This help
    -v  Show verison
```
## Super simple to use
RunnerGo is designed to be the simplest way possible to make stress test. 

```json
{
    "c": 2,
    "n": 2,
    "data": {
        "method": "POST",
        "url": "http://echo.apipost.com/get.php",
        "mode": "urlencoded",
        "headers": [
            {
                "name": "Pragma",
                "value": "no-cache"
            },
            {
                "name": "Server",
                "value": "yisu.com"
            }
        ],
        "postData": {
            "params": [
                {
                    "name": "logo",
                    "type": "file",
                    "value": "@/Users/mhw/Downloads/"
                },
                {
                    "name": "title",
                    "value": "标题"
                }
            ]
        }
    }
}
```

## Examples
```json
{
    "code": 200,
    "message": "success",
    "data": {
        "CompleteRequests": 4,
        "FailedRequests": 0,
        "SuccessRequests": 4,
        "TimeToken": 0.324641,
        "TotalDataSize": 2028,
        "AvgDataSize": 507,
        "RequestsPerSec": 12.32130260811173,
        "SuccessRequestsPerSec": 12.32130260811173,
        "MinUseTime": 60,
        "MaxUseTime": 249,
        "AvgUseTime": 156.25,
        "CodeDetail": {
            "200": 4
        },
        "WaitingTimeDetail": {
            "10.00%": 70,
            "25.00%": 70,
            "50.00%": 74,
            "75.00%": 232,
            "90.00%": 232,
            "95.00%": 232,
            "99.00%": 232,
            "99.90%": 232,
            "99.99%": 232
        },
        "AvgConn": 77.5,
        "MaxConn": 158,
        "MinConn": 0,
        "AvgDNS": 30.5,
        "MaxDNS": 61,
        "MinDNS": 0,
        "AvgReq": 0,
        "MaxReq": 0,
        "MinReq": 0,
        "AvgDelay": 77.5,
        "MaxDelay": 90,
        "MinDelay": 60,
        "AvgRes": 0,
        "MaxRes": 0,
        "MinRes": 0
    }
}
```