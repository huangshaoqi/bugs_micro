FROM golang:1.19-alpine3.17 AS builder
WORKDIR /go_work/notificator
ADD notificator .

RUN go env -w  GOPROXY="https://goproxy.cn,direct"  \
    && go env -w  GO111MODULE=on  \
    && go env -w CGO_ENABLED=0 \
    && go mod tidy \
    && go build -o ./notificator ./cmd/
# ENTRYPOINT ./notificator


FROM alpine:3.17.2
WORKDIR /go_work/notificator
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk add --no-cache tzdata

COPY --from=builder /go_work/notificator/notificator /go_work/notificator/
ENTRYPOINT ./notificator
