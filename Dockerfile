# 编译website
FROM golang:1.14.9-alpine3.12 AS builder

ADD . /go/src/github.com/eudore/website
RUN apk add git && \
	go version && go env && \
	for i in $(go build /go/src/github.com/eudore/website/app.go 2>&1 | grep find | cut -d\" -f2);do GO111MODULE=off go get -v $i; done && \
	mkdir website && \
	cp -r /go/src/github.com/eudore/website/static website/static && \
	cp -r /go/src/github.com/eudore/website/config website/config && \
	GO111MODULE=off CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags "-s -w -X main.BuildTime=`date '+%Y-%m-%d_%I:%M:%S'` -X main.CommitID=`git --git-dir=/go/src/github.com/eudore/website/.git rev-parse HEAD`"  -o website/server /go/src/github.com/eudore/website/app.go


# 创建运行镜像
FROM alpine:3.12

COPY --from=builder /go/website /

CMD ["/server"]
