package main

import (
	"fmt"

	service "github.com/huangshaoqi/bugs_micro/users/cmd/service"
)

func main() {
	service.Run()
	fmt.Println("test users 1")
}
