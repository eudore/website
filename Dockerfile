# 编译website
FROM golang:1.13.0-alpine3.10 AS builder

WORKDIR /go/src

RUN apk add git
RUN go version && go env
RUN GO111MODULE=off go get -v github.com/eudore/website
RUN GO111MODULE=off CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags "-s -w" -o server github.com/eudore/website


# 创建运行镜像
FROM alpine:latest

COPY --from=builder /go/src/server /
COPY --from=builder /go/src/github.com/eudore/website/config/config.json /config.json
COPY --from=builder /go/src/github.com/eudore/website/static /static

CMD ["/server", "--enable.+=docker"]
