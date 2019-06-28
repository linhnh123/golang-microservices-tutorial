package main

import (
	"flag"
	"fmt"

	"github.com/linhnh123/golang-microservices-tutorial/common/messaging"

	"github.com/linhnh123/golang-microservices-tutorial/accountservice/config"

	"github.com/spf13/viper"

	"github.com/linhnh123/golang-microservices-tutorial/accountservice/dbclient"

	"github.com/linhnh123/golang-microservices-tutorial/accountservice/service"

	commonConfig "github.com/linhnh123/golang-microservices-tutorial/common/config"
)

var appName = "accountservice"

func initializeBoltClient() {
	service.DBClient = &dbclient.BoltClient{}
	service.DBClient.OpenBoltDb()
	service.DBClient.Seed()
	service.DBClient.CloseBoltDb()
}

func initializeMessaging() {
	if !viper.IsSet("amqp_server_url") {
		panic("Not set 'amqp_server_url'")
	}
	service.MessagingClient = &messaging.MessagingClient{}
	service.MessagingClient.ConnectToBroker(viper.GetString("amqp_server_url"))
	service.MessagingClient.Subscribe(viper.GetString("config_event_bus"), "topic", appName, commonConfig.HandleRefreshEvent)
}

func init() {
	profile := flag.String("profile", "test", "Environment profile")
	configServerUrl := flag.String("configServerUrl", "http://configserver:8888", "Address to config server")
	configBranch := flag.String("configBranch", "master", "git branch to fetch configuration from")

	flag.Parse()

	viper.Set("profile", *profile)
	viper.Set("configServerUrl", *configServerUrl)
	viper.Set("configBranch", *configBranch)
}

func main() {
	fmt.Printf("Starting %v\n", appName)

	config.LoadConfigurationFromBranch(
		viper.GetString("configServerUrl"),
		appName,
		viper.GetString("profile"),
		viper.GetString("configBranch"),
	)

	initializeBoltClient()
	initializeMessaging()
	go config.StartListener(appName, viper.GetString("amqp_server_url"), viper.GetString("config_event_bus"))
	service.StartWebServer(viper.GetString("server_port"))
}
