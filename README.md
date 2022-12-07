

RunnerGo is a develop tool similar to apache bench (ab).

## Usage
* runnerGo -h

```
Options:
    -n 	requests     Number of requests to perform
	-c 	concurrency  Number of multiple requests to make at a time
	-t  Timeout for each request in seconds, Default is 10
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
        "url": "https://echo.apipost.cn/get.php",
        "httpVersion": "HTTP/1.1",
        "mode": "urlencoded",
        "headers": [],
        "queryString": [],
        "cookies": [],
        "headersSize": 670,
        "bodySize": 279,
        "postData": {
            "mimeType": "multipart/form-data; boundary=----WebKitFormBoundaryt1AKSW2uGI9p3PPS",
            "text": "------WebKitFormBoundaryt1AKSW2uGI9p3PPS\r\nContent-Disposition: form-data; name=\"logo\"; filename=\"har.json\"\r\nContent-Type: text/x-sh\r\n\r\n\r\n------WebKitFormBoundaryt1AKSW2uGI9p3PPS\r\nContent-Disposition: form-data; name=\"title\"\r\n\r\n标题\r\n------WebKitFormBoundaryt1AKSW2uGI9p3PPS--\r\n",
            "params": [
                {
                    "name": "logo",
                    "type": "file",
                    "value": "@/Users/mhw/Downloads/har.json"
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