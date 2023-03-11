FROM golang:1.19
WORKDIR /go_work/users

ADD users .

RUN go env -w GOPROXY="https://goproxy.cn,direct"
RUN go env -w GO111MODULE=auto

RUN go install github.com/radovskyb/watcher/cmd/watcher@latest && go mod tidy

ENTRYPOINT  go run ./cmd && watcher -cmd='ps -ef | grep "go run ./cmd" | grep -v grep | awk '{print $2}' | xargs kill -9 && go run ./cmd/'
