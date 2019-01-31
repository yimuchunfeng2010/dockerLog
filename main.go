package main

import (
	"DockerLog/service"
)

func init(){
	service.GetAllService()
}
func main() {
	service.LogMain()
}
