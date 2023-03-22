FROM golang:1.19 AS builder
WORKDIR /go_work/bugs
ADD bugs .

RUN go env -w  GOPROXY="https://goproxy.cn,direct"  \
    && go env -w  GO111MODULE=on  \
    && go env -w CGO_ENABLED=0 \
    && go mod tidy \
    && go build -o ./bugs ./cmd/
# ENTRYPOINT ./bugs


FROM alpine:3.17.2
WORKDIR /go_work/bugs
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk add --no-cache tzdata

COPY --from=builder /go_work/bugs/bugs /go_work/bugs/
ENTRYPOINT ./bugs