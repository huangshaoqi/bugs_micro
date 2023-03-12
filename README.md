## 使用 GO-KIT 来构建微服务

### 创建一个 bug 微服务追踪系统：

一共是以下三个微服务模块

- Users
- Bugs
- Notificator

### 设置工作区（方便多模块依赖）

```bash
mkdir bugs_micro #总根目录
go mod init github.com/huangshaoqi/bugs_micro
go work init users bugs notificator
```

### go-kit CLI

创建独立的包可从模板创建服务:

```bash
go get github.com/go-kit/kit
go install github.com/GrantZheng/kit@latest
```

创建以下服务:

```
kit new service users --module github.com/huangshaoqi/bugs_micro/users # 简写: kit n s users -m github.com/huangshaoqi/bugs_micro/users
kit new service bugs -m github.com/huangshaoqi/bugs_micro/bugs
kit new service notificator -m github.com/huangshaoqi/bugs_micro/notificator
```

这些命令将生成初始的文件夹结构和服务接口。接口默认为空，让我们定义接口中的函数。从创建用户函数开始。

users:

```golang
package service

import "context"

// 用户服务
type UsersService interface {
    Create(ctx context.Context, email string) error
}
```

bugs

```golang
package service

import "context"

// bug服务
type BugsService interface {
    // 在这里添加你自己的方法
    Create(ctx context.Context, bug string) error
}
```

notifcator:

```golang
package service

import "context"

// 通知服务
type NotificatorService interface {
    // 在这里添加你自己的方法
    SendEmail(ctx context.Context, email string, content string) (string,error)
}
```

然后我们需要运行一个命令生成一个服务，这个命令将创建服务模板，服务中间件和终端代码。同时它还创建了一个 cmd/ 包来运行我们的服务。

```bash
kit generate service users --dmw # 简写:kit g s users --dmw
kit generate service bugs --dmw
```

-dmw 参数创建默认的终端中间件和日志中间件。

这个命令已经将 go-kit 工具的端点包和 HTTP 传输包添加到我们的代码中。现在我们只需要在任意一个位置实现我们的业务代码就可以了。
因为通知服务是一个内部服务，所以不需要 REST API，我们用 gRPC 来实现它。gRPC 是谷歌的 RPC 框架，如果你以前从未使用过，请看这里。

为此，我们需要先安装 protoc 和 protobuf 。

```bash
kit generate service notificator -t grpc --dmw
```

这同时生成了 .pb 文件，我们将在下一篇文章中去解决它。
go-kit CLI 还可以创建 docker-compose 模板，让我们试一试。

```bash
kit generate docker
```

这个命令生成了 Dockerfile 和带有端口映射的 docker-compose.yml 。将它运行起来并调用 /create 接口 。

```bash
docker-compose up
```

Dockerfiles 使用了 go 的 watcher 包，如果 Go 代码发生了变化，它会更新并重启二进制文件，这在本地环境中非常方便。
注意:需要修改 notificator/pkg/grpc/pb/complie.sh
把 notificator 加上后缀.proto，其他 2 个服务也一样要改

```bash
protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative notificator.proto
```

现在我们的服务运行的端口是 8800，8801，8802。可以试着调用用户服务

```
curl -XPOST http://localhost:8800/create -d '{"email": "test"}'
```

### 实现 Notificator 服务

修改 notificator/pkg/grpc/pb/notificator.proto

```proto3
syntax = "proto3";

package pb;

option go_package = "github.com/huangshaoqi/bugs_micro/notificator/pkg/grpc/pb";


//The Notificator service definition.
service Notificator {
 rpc SendEmail (SendEmailRequest) returns (SendEmailReply);
}
message SendEmailRequest {
  string email = 1;
  string content = 2;
}
message SendEmailReply {
  string id = 1;
}
```

现在我们要生成服务端和客户端的存根，可以使用 kit 工具生成的 compile.sh 脚本，脚本已经包含 protoc 命令。

```bash
cd notificator/pkg/grpc/pb
./compile.sh
```

可以发现，notificator.pb.go 这个文件已经更新了。
现在我们需要实现服务本身了。我们用生成一个 UUID 代替发送真实的电子邮件。首先需要对服务进行调整以匹配我们的请求和响应格式（返回一个新的参数：id）。

notificator/pkg/service/service.go:

```golang
package service

import (
	"context"
	uuid "github.com/satori/go.uuid"
)

// NotificatorService describes the service.
type NotificatorService interface {
	// Add your methods here
	// e.x: Foo(ctx context.Context,s string)(rs string, err error)

	SendEmail(ctx context.Context, email string, content string) (string, error)
}

type basicNotificatorService struct{}

func (b *basicNotificatorService) SendEmail(ctx context.Context, email string, content string) (string, error) {
	// TODO implement the business logic of SendEmail
	uuidObjet := uuid.NewV4()
	v, err := uuidObjet.Value()
	if err != nil {
		return "", err
	}
	return v.(string), nil
}

// NewBasicNotificatorService returns a naive, stateless implementation of NotificatorService.
func NewBasicNotificatorService() NotificatorService {
	return &basicNotificatorService{}
}

// New returns a NotificatorService with all of the expected middleware wired in.
func New(middleware []Middleware) NotificatorService {
	var svc NotificatorService = NewBasicNotificatorService()
	for _, m := range middleware {
		svc = m(svc)
	}
	return svc
}
```

notificator/pkg/service/middleware.go:

```golang
package service

import (
	"context"
	log "github.com/go-kit/kit/log"
)

// Middleware describes a service middleware.
type Middleware func(NotificatorService) NotificatorService

type loggingMiddleware struct {
	logger log.Logger
	next   NotificatorService
}

// LoggingMiddleware takes a logger as a dependency
// and returns a NotificatorService Middleware.
func LoggingMiddleware(logger log.Logger) Middleware {
	return func(next NotificatorService) NotificatorService {
		return &loggingMiddleware{logger, next}
	}

}

func (l loggingMiddleware) SendEmail(ctx context.Context, email string, content string) (string, error) {
	defer func() {
		l.logger.Log("method", "SendEmail", "email", email, "content", content)
	}()
	return l.next.SendEmail(ctx, email, content)
}

```

notificator/pkg/endpoint/endpoint.go

```golang
package endpoint

import (
	"context"
	endpoint "github.com/go-kit/kit/endpoint"
	service "github.com/huangshaoqi/bugs_micro/notificator/pkg/service"
)

// SendEmailRequest collects the request parameters for the SendEmail method.
type SendEmailRequest struct {
	Email   string `json:"email"`
	Content string `json:"content"`
}

// SendEmailResponse collects the response parameters for the SendEmail method.
type SendEmailResponse struct {
	Id string
	E0 error `json:"e0"`
}

// MakeSendEmailEndpoint returns an endpoint that invokes SendEmail on the service.
func MakeSendEmailEndpoint(s service.NotificatorService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(SendEmailRequest)
		id, e0 := s.SendEmail(ctx, req.Email, req.Content)
		return SendEmailResponse{Id: id, E0: e0}, nil
	}
}

// Failed implements Failer.
func (r SendEmailResponse) Failed() error {
	return r.E0
}

// Failure is an interface that should be implemented by response types.
// Response encoders can check if responses are Failer, and if so they've
// failed, and if so encode them using a separate write path based on the error.
type Failure interface {
	Failed() error
}

// SendEmail implements Service. Primarily useful in a client.
func (e Endpoints) SendEmail(ctx context.Context, email string, content string) (e0 error) {
	request := SendEmailRequest{
		Content: content,
		Email:   email,
	}
	response, err := e.SendEmailEndpoint(ctx, request)
	if err != nil {
		return
	}
	return response.(SendEmailResponse).E0
}

```

如果我们搜索 TODO grep -R "TODO" notificator ，可以看到还需要实现 gRPC 请求和响应的编码和解码。
notificator/pkg/grpc/handler.go:

````golang
package grpc

import (
	"context"
	grpc "github.com/go-kit/kit/transport/grpc"
	endpoint "github.com/huangshaoqi/bugs_micro/notificator/pkg/endpoint"
	pb "github.com/huangshaoqi/bugs_micro/notificator/pkg/grpc/pb"
)

// makeSendEmailHandler creates the handler logic
func makeSendEmailHandler(endpoints endpoint.Endpoints, options []grpc.ServerOption) grpc.Handler {
	return grpc.NewServer(endpoints.SendEmailEndpoint, decodeSendEmailRequest, encodeSendEmailResponse, options...)
}

// decodeSendEmailResponse is a transport/grpc.DecodeRequestFunc that converts a
// gRPC request to a user-domain SendEmail request.
// TODO implement the decoder
func decodeSendEmailRequest(_ context.Context, r interface{}) (interface{}, error) {
	req := r.(*pb.SendEmailRequest)
	return endpoint.SendEmailRequest{Email: req.Email, Content: req.Content}, nil
}

// encodeSendEmailResponse is a transport/grpc.EncodeResponseFunc that converts
// a user-domain response to a gRPC reply.
// TODO implement the encoder
func encodeSendEmailResponse(_ context.Context, r interface{}) (interface{}, error) {
	reply := r.(endpoint.SendEmailResponse)
	return &pb.SendEmailReply{Id: reply.Id}, nil
}
func (g *grpcServer) SendEmail(ctx context.Context, req *pb.SendEmailRequest) (*pb.SendEmailReply, error) {
	_, rep, err := g.sendEmail.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*pb.SendEmailReply), nil
}

```
服务发现
SendEmail 接口将由用户服务调用，所以用户服务必须知道通知服务的地址，这是典型的服务发现问题。在本地环境中，我们使用 Docker Compose 部署时知道如何连接服务，但是，在实际的分布式环境中，可能会遇到其他问题。

首先，在 etcd 中注册我们的通知服务。etcd 是一种可靠的分布式键值存储，广泛应用于服务发现。go-kit 支持使用其他的服务发现技术如：eureka、consul 和 zookeeper 等。

把它添加进 Docker Compose 中，这样我们的服务就可以使用它了。CV 工程师上线：
bugs_micro/docker-compose.yml:

```YAML
version: "3"

networks:
  etcd-net: # 网络
    driver: bridge    # 桥接模式
volumes:

  etcd1_data:         # 挂载到本地的数据卷名
    driver: local
  etcd2_data:
    driver: local
  etcd3_data:
    driver: local
  users:
    driver: local
  bugs:
    driver: local
  notificator:
    driver: local

services:
  etcd1:
    image: bitnami/etcd:latest  # 镜像
    container_name: etcd1       # 容器名 --name
    restart: always             # 总是重启
    networks:
      - etcd-net                # 使用的网络 --network
    ports: # 端口映射 -p
      - "20000:2379"
      - "20001:2380"
    environment: # 环境变量 --env
      - ALLOW_NONE_AUTHENTICATION=yes                       # 允许不用密码登录
      - ETCD_NAME=etcd1                                     # etcd 的名字
      - ETCD_INITIAL_ADVERTISE_PEER_URLS=http://etcd1:2380  # 列出这个成员的伙伴 URL 以便通告给集群的其他成员
      - ETCD_LISTEN_PEER_URLS=http://0.0.0.0:2380           # 用于监听伙伴通讯的URL列表
      - ETCD_LISTEN_CLIENT_URLS=http://0.0.0.0:2379         # 用于监听客户端通讯的URL列表
      - ETCD_ADVERTISE_CLIENT_URLS=http://etcd1:2379        # 列出这个成员的客户端URL，通告给集群中的其他成员
      - ETCD_INITIAL_CLUSTER_TOKEN=etcd-cluster             # 在启动期间用于 etcd 集群的初始化集群记号
      - ETCD_INITIAL_CLUSTER=etcd1=http://etcd1:2380,etcd2=http://etcd2:2380,etcd3=http://etcd3:2380 # 为启动初始化集群配置
      - ETCD_INITIAL_CLUSTER_STATE=new                      # 初始化集群状态
    volumes:
      - /Users/huangshaoqi/Data/gowork/2023/bugs_micro/etcd1_data:/bitnami/etcd                            # 挂载的数据卷

  etcd2:
    image: bitnami/etcd:latest
    container_name: etcd2
    restart: always
    networks:
      - etcd-net
    ports:
      - "20002:2379"
      - "20003:2380"
    environment:
      - ALLOW_NONE_AUTHENTICATION=yes
      - ETCD_NAME=etcd2
      - ETCD_INITIAL_ADVERTISE_PEER_URLS=http://etcd2:2380
      - ETCD_LISTEN_PEER_URLS=http://0.0.0.0:2380
      - ETCD_LISTEN_CLIENT_URLS=http://0.0.0.0:2379
      - ETCD_ADVERTISE_CLIENT_URLS=http://etcd2:2379
      - ETCD_INITIAL_CLUSTER_TOKEN=etcd-cluster
      - ETCD_INITIAL_CLUSTER=etcd1=http://etcd1:2380,etcd2=http://etcd2:2380,etcd3=http://etcd3:2380
      - ETCD_INITIAL_CLUSTER_STATE=new
    volumes:
      - /Users/huangshaoqi/Data/gowork/2023/bugs_micro/etcd2_data:/bitnami/etcd

  etcd3:
    image: bitnami/etcd:latest
    container_name: etcd3
    restart: always
    networks:
      - etcd-net
    ports:
      - "20004:2379"
      - "20005:2380"
    environment:
      - ALLOW_NONE_AUTHENTICATION=yes
      - ETCD_NAME=etcd3
      - ETCD_INITIAL_ADVERTISE_PEER_URLS=http://etcd3:2380
      - ETCD_LISTEN_PEER_URLS=http://0.0.0.0:2380
      - ETCD_LISTEN_CLIENT_URLS=http://0.0.0.0:2379
      - ETCD_ADVERTISE_CLIENT_URLS=http://etcd3:2379
      - ETCD_INITIAL_CLUSTER_TOKEN=etcd-cluster
      - ETCD_INITIAL_CLUSTER=etcd1=http://etcd1:2380,etcd2=http://etcd2:2380,etcd3=http://etcd3:2380
      - ETCD_INITIAL_CLUSTER_STATE=new
    volumes:
      - /Users/huangshaoqi/Data/gowork/2023/bugs_micro/etcd3_data:/bitnami/etcd

  ### etcd 其他环境配置见
  bugs:
    networks:
      - etcd-net
    build:
      context: .
      dockerfile: bugs/Dockerfile
    container_name: bugs
    ports:
    - 8800:8081
    restart: always
    volumes:
    - /Users/huangshaoqi/Data/gowork/2023/bugs_micro/bugs:/go_work/bugs
  notificator:
    networks:
      - etcd-net
    build:
      context: .
      dockerfile: notificator/Dockerfile
    container_name: notificator
    ports:
    - 8802:8082
    restart: always
    volumes:
    - /Users/huangshaoqi/Data/gowork/2023/bugs_micro/notificator:/go_work/notificator
  users:
    networks:
      - etcd-net
    build:
      context: .
      dockerfile: users/Dockerfile
    container_name: users
    ports:
    - 8801:8081
    restart: always
    volumes:
    - /Users/huangshaoqi/Data/gowork/2023/bugs_micro/users:/go_work/users
````

在 etcd 中注册通知服务 notificator/cmd/service/service.go：

```golang
package service

import (
	"context"
	"flag"
	"fmt"
	endpoint1 "github.com/go-kit/kit/endpoint"
	log "github.com/go-kit/kit/log"
	prometheus "github.com/go-kit/kit/metrics/prometheus"
	sdetcd "github.com/go-kit/kit/sd/etcdv3"
	endpoint "github.com/huangshaoqi/bugs_micro/notificator/pkg/endpoint"
	grpc "github.com/huangshaoqi/bugs_micro/notificator/pkg/grpc"
	pb "github.com/huangshaoqi/bugs_micro/notificator/pkg/grpc/pb"
	service "github.com/huangshaoqi/bugs_micro/notificator/pkg/service"
	lightsteptracergo "github.com/lightstep/lightstep-tracer-go"
	group "github.com/oklog/oklog/pkg/group"
	opentracinggo "github.com/opentracing/opentracing-go"
	zipkingoopentracing "github.com/openzipkin-contrib/zipkin-go-opentracing"
	zipkingo "github.com/openzipkin/zipkin-go"
	http "github.com/openzipkin/zipkin-go/reporter/http"
	prometheus1 "github.com/prometheus/client_golang/prometheus"
	promhttp "github.com/prometheus/client_golang/prometheus/promhttp"
	grpc1 "google.golang.org/grpc"
	"net"
	http1 "net/http"
	"os"
	"os/signal"
	appdash "sourcegraph.com/sourcegraph/appdash"
	opentracing "sourcegraph.com/sourcegraph/appdash/opentracing"
	"syscall"
)

var tracer opentracinggo.Tracer
var logger log.Logger

// Define our flags. Your service probably won't need to bind listeners for
// all* supported transports, but we do it here for demonstration purposes.
var fs = flag.NewFlagSet("notificator", flag.ExitOnError)
var debugAddr = fs.String("debug-addr", ":8080", "Debug and metrics listen address")
var httpAddr = fs.String("http-addr", ":8081", "HTTP listen address")
var grpcAddr = fs.String("grpc-addr", ":8082", "gRPC listen address")
var thriftAddr = fs.String("thrift-addr", ":8083", "Thrift listen address")
var thriftProtocol = fs.String("thrift-protocol", "binary", "binary, compact, json, simplejson")
var thriftBuffer = fs.Int("thrift-buffer", 0, "0 for unbuffered")
var thriftFramed = fs.Bool("thrift-framed", false, "true to enable framing")
var zipkinURL = fs.String("zipkin-url", "", "Enable Zipkin tracing via a collector URL e.g. http://localhost:9411/api/v1/spans")
var lightstepToken = fs.String("lightstep-token", "", "Enable LightStep tracing via a LightStep access token")
var appdashAddr = fs.String("appdash-addr", "", "Enable Appdash tracing via an Appdash server host:port")

func Run() {
	fs.Parse(os.Args[1:])

	// Create a single logger, which we'll use and give to other components.
	logger = log.NewLogfmtLogger(os.Stderr)
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	logger = log.With(logger, "caller", log.DefaultCaller)

	//  Determine which tracer to use. We'll pass the tracer to all the
	// components that use it, as a dependency
	if *zipkinURL != "" {
		logger.Log("tracer", "Zipkin", "URL", *zipkinURL)
		reporter := http.NewReporter(*zipkinURL)
		defer reporter.Close()
		endpoint, err := zipkingo.NewEndpoint("notificator", "localhost:80")
		if err != nil {
			logger.Log("err", err)
			os.Exit(1)
		}
		localEndpoint := zipkingo.WithLocalEndpoint(endpoint)
		nativeTracer, err := zipkingo.NewTracer(reporter, localEndpoint)
		if err != nil {
			logger.Log("err", err)
			os.Exit(1)
		}
		tracer = zipkingoopentracing.Wrap(nativeTracer)
	} else if *lightstepToken != "" {
		logger.Log("tracer", "LightStep")
		tracer = lightsteptracergo.NewTracer(lightsteptracergo.Options{AccessToken: *lightstepToken})
		defer lightsteptracergo.Flush(context.Background(), tracer)
	} else if *appdashAddr != "" {
		logger.Log("tracer", "Appdash", "addr", *appdashAddr)
		collector := appdash.NewRemoteCollector(*appdashAddr)
		tracer = opentracing.NewTracer(collector)
		defer collector.Close()
	} else {
		logger.Log("tracer", "none")
		tracer = opentracinggo.GlobalTracer()
	}

	svc := service.New(getServiceMiddleware(logger))
	eps := endpoint.New(svc, getEndpointMiddleware(logger))
	g := createService(eps)
	initMetricsEndpoint(g)
	initCancelInterrupt(g)

    // TODO 执行注册中心函数
	registrar, err := registerService(logger)
	if err != nil {
		logger.Log(err)
		return
	}
	defer registrar.Deregister()
	logger.Log("exit", g.Run())

}
// TODO 注册中心函数
func registerService(logger log.Logger) (*sdetcd.Registrar, error) {
	var (
		etcdServers = []string{"http://etcd1:2379", "http://etcd2:2379", "http://etcd3:2379"}
		prefix      = "/services/notificator/"
		instance    = "notificator:8082"
		key         = prefix + instance
	)
	client, err := sdetcd.NewClient(context.Background(), etcdServers, sdetcd.ClientOptions{})
	if err != nil {
		return nil, err
	}
	register := sdetcd.NewRegistrar(client, sdetcd.Service{
		Key:   key,
		Value: instance,
	}, logger)

	register.Register()
	return register, nil
}

func initGRPCHandler(endpoints endpoint.Endpoints, g *group.Group) {
	options := defaultGRPCOptions(logger, tracer)
	// Add your GRPC options here

	grpcServer := grpc.NewGRPCServer(endpoints, options)
	grpcListener, err := net.Listen("tcp", *grpcAddr)
	if err != nil {
		logger.Log("transport", "gRPC", "during", "Listen", "err", err)
	}
	g.Add(func() error {
		logger.Log("transport", "gRPC", "addr", *grpcAddr)
		baseServer := grpc1.NewServer()
		pb.RegisterNotificatorServer(baseServer, grpcServer)
		return baseServer.Serve(grpcListener)
	}, func(error) {
		grpcListener.Close()
	})

}
func getServiceMiddleware(logger log.Logger) (mw []service.Middleware) {
	mw = []service.Middleware{}
	mw = addDefaultServiceMiddleware(logger, mw)
	// Append your middleware here

	return
}
func getEndpointMiddleware(logger log.Logger) (mw map[string][]endpoint1.Middleware) {
	mw = map[string][]endpoint1.Middleware{}
	duration := prometheus.NewSummaryFrom(prometheus1.SummaryOpts{
		Help:      "Request duration in seconds.",
		Name:      "request_duration_seconds",
		Namespace: "example",
		Subsystem: "notificator",
	}, []string{"method", "success"})
	addDefaultEndpointMiddleware(logger, duration, mw)
	// Add you endpoint middleware here

	return
}
func initMetricsEndpoint(g *group.Group) {
	http1.DefaultServeMux.Handle("/metrics", promhttp.Handler())
	debugListener, err := net.Listen("tcp", *debugAddr)
	if err != nil {
		logger.Log("transport", "debug/HTTP", "during", "Listen", "err", err)
	}
	g.Add(func() error {
		logger.Log("transport", "debug/HTTP", "addr", *debugAddr)
		return http1.Serve(debugListener, http1.DefaultServeMux)
	}, func(error) {
		debugListener.Close()
	})
}
func initCancelInterrupt(g *group.Group) {
	cancelInterrupt := make(chan struct{})
	g.Add(func() error {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		select {
		case sig := <-c:
			return fmt.Errorf("received signal %s", sig)
		case <-cancelInterrupt:
			return nil
		}
	}, func(error) {
		close(cancelInterrupt)
	})
}

```

现在我们使用用户服务调用通知服务，当我们创建一个服务后它会发送一个虚构的通知给用户。

由于通知服务是一个 gRPC 服务，在用户服务中，我们需要和客户端共用一个客户端存根。

protobuf 的客户端存根代码在 notificator/pkg/grpc/pb/notificator.pb.go ，我们把这个包引入我们的客户端

users/pkg/service/service.go:

```golang
package service

import (
	"context"
	sdetcd "github.com/go-kit/kit/sd/etcdv3"
	"github.com/huangshaoqi/bugs_micro/notificator/pkg/grpc/pb"
	"google.golang.org/grpc"
	"log"
)

// UsersService describes the service.
type UsersService interface {
	// Add your methods here
	// e.x: Foo(ctx context.Context,s string)(rs string, err error)

	Create(ctx context.Context, email string) error
}

type basicUsersService struct {
	notificatorServiceClient pb.NotificatorClient
}

func (b *basicUsersService) Create(ctx context.Context, email string) (e0 error) {
	// TODO implement the business logic of Create
	reply, e0 := b.notificatorServiceClient.SendEmail(context.Background(),
		&pb.SendEmailRequest{
			Email:   email,
			Content: "hi,this is test content",
		})
	if reply != nil {
		log.Printf("Email Id: %s", reply.Id)
	}
	return e0
}

// NewBasicUsersService returns a naive, stateless implementation of UsersService.
func NewBasicUsersService() UsersService {
	/*
		conn, err := grpc.Dial("notificator:8082", grpc.WithInsecure())
		if err != nil {
			log.Printf("unable to connect to notificator: %s", err.Error())
			return new(basicUsersService)
		}
		log.Printf("connected to notificator")
		return &basicUsersService{
			notificatorServiceClient: pb.NewNotificatorClient(conn),
		}
	*/
	var etcdServers = []string{"http://etcd1:2379", "http://etcd2:2379", "http://etcd3:2379"}
	client, err := sdetcd.NewClient(context.Background(), etcdServers, sdetcd.ClientOptions{})
	if err != nil {
		log.Printf("unable to connect to etcd: %s", err.Error())
		return new(basicUsersService)
	}
	entries, err := client.GetEntries("/services/notificator")
	if err != nil || len(entries) == 0 {
		log.Printf("unable to get prefix entries: %s", err.Error())
		return new(basicUsersService)
	}
	conn, err := grpc.Dial(entries[0], grpc.WithInsecure())
	if err != nil {
		log.Printf("unable to connect to notificator: %s", err.Error())
		return new(basicUsersService)
	}
	log.Printf("connected to notificator")
	return &basicUsersService{
		notificatorServiceClient: pb.NewNotificatorClient(conn),
	}
}

// New returns a UsersService with all of the expected middleware wired in.
func New(middleware []Middleware) UsersService {
	var svc UsersService = NewBasicUsersService()
	for _, m := range middleware {
		svc = m(svc)
	}
	return svc
}

```

因为我们只有一个服务，所以只获取到了一个连接，但实际系统中可能有上百个服务，所以我们可以应用一些逻辑来进行实例选择，例如轮询。
现在让我们启动用户服务来测试一下：

```
cd bugs_micro
docker-compose up
```
