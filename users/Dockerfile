FROM golang:1.19-alpine3.17 AS builder
WORKDIR /go_work/users
ADD users .

RUN go env -w  GOPROXY="https://goproxy.cn,direct"  \
    && go env -w  GO111MODULE=on  \
    && go env -w CGO_ENABLED=0 \
    && go mod tidy \
    && go build -o ./users ./cmd/
# ENTRYPOINT ./users


FROM alpine:3.17.2
WORKDIR /go_work/users
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk add --no-cache tzdata

COPY --from=builder /go_work/users/users /go_work/users/
ENTRYPOINT ./users
