package main

import (
	"fmt"
	service "notificator/cmd/service"
)

func main() {
	service.Run()
	fmt.Println("test notificator")
}
