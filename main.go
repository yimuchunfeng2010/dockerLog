package main

import (
	"DockerLog/service"
)

func init(){
	service.InitServices()
}
func main() {
	service.LogMain()
}
