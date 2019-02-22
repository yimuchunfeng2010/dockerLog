package main

import (
	"dockerLog/service"
)

func init() {
	service.InitServices()
}
func main() {
	service.LogMain()
}
