# 编译website
FROM golang:1.13.0-alpine3.10 AS builder

RUN apk add git && \
	go version && go env && \
	GO111MODULE=off go get -v github.com/eudore/website && \
	mkdir website && cp -r /go/src/github.com/eudore/website/static website/static && \
	GO111MODULE=off CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags "-s -w" -o website/server github.com/eudore/website


# 创建运行镜像
FROM alpine:latest

COPY --from=builder /go/website /go/src/github.com/eudore/website/config/config.json /

CMD ["/server"]
