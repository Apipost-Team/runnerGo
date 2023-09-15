# 打包依赖阶段使用golang作为基础镜像
FROM golang:1.19 AS build-stage
#ENV GOPROXY=https://goproxy.cn,direct
WORKDIR /app

COPY . .

#下载依赖
RUN go mod download
# CGO_ENABLED禁用cgo 然后指定OS等，并go build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o runnerGo_linux_x64 runnerGo.go


# 运行阶段指定scratch作为基础镜像
FROM runnergo/debian:stable-slim

WORKDIR /app

# 将上一个阶段编译程序copy到app目录
COPY --from=build-stage /app/runnerGo_linux_x64 .

EXPOSE 10397

ENTRYPOINT ["/app/runnerGo_linux_x64"]