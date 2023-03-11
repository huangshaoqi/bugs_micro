package main

import (
	service "bugs/cmd/service"
	"fmt"
)

func main() {
	service.Run()
	fmt.Println("test bugs")
}
