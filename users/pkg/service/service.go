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
