package main

import (
	"fmt"
	service "users/cmd/service"
)

func main() {
	service.Run()
	fmt.Println("test users")
}
