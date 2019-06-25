package main

import (
	"fmt"

	"github.com/linhnh123/golang-microservices-tutorial/accountservice/dbclient"

	"github.com/linhnh123/golang-microservices-tutorial/accountservice/service"
)

var appName = "accountservice"

func initializeBoltClient() {
	service.DBClient = &dbclient.BoltClient{}
	service.DBClient.OpenBoltDb()
	service.DBClient.Seed()
	service.DBClient.CloseBoltDb()
}

func main() {
	fmt.Printf("Starting %v\n", appName)
	initializeBoltClient()
	service.StartWebServer("6767")
}
