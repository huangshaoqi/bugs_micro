package main

import (
	"fmt"
	service "github.com/huangshaoqi/bugs_micro/notificator/cmd/service"
)

func main() {
	service.Run()
	fmt.Println("test notificator 1")
}
