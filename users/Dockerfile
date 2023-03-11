FROM golang:1.19
WORKDIR /go_work/users

ADD users .

RUN go env -w GOPROXY="https://goproxy.cn,direct"
RUN go env -w GO111MODULE=auto

RUN go install github.com/radovskyb/watcher/cmd/watcher@latest && go mod tidy

ENTRYPOINT  ./start.sh && watcher -ignore "log.txt" -ignore "nohup.out" -cmd='./restart.sh'
