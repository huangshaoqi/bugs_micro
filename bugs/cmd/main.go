package main

import (
	"fmt"
	service "github.com/huangshaoqi/bugs_micro/bugs/cmd/service"
)

func main() {
	service.Run()
	fmt.Println("test bugs")
}
