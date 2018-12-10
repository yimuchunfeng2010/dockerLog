package main

import (
	"log/service"
)

func init(){
	service.GetAllService()
}
func main() {
	service.LogMain()
}
