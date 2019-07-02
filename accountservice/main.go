package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/linhnh123/golang-microservices-tutorial/common/messaging"
	"github.com/linhnh123/golang-microservices-tutorial/common/tracing"
	"github.com/sirupsen/logrus"

	"github.com/spf13/viper"

	"github.com/linhnh123/golang-microservices-tutorial/accountservice/dbclient"

	"github.com/linhnh123/golang-microservices-tutorial/accountservice/service"

	"github.com/linhnh123/golang-microservices-tutorial/common/config"

	cb "github.com/linhnh123/golang-microservices-tutorial/common/circuitbreaker"
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
	service.MessagingClient.Subscribe(viper.GetString("config_event_bus"), "topic", appName, config.HandleRefreshEvent)
}

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02T15:04:05.000",
		FullTimestamp:   true,
	})
	profile := flag.String("profile", "test", "Environment profile")
	configServerUrl := flag.String("configServerUrl", "http://configserver:8888", "Address to config server")
	configBranch := flag.String("configBranch", "master", "git branch to fetch configuration from")

	flag.Parse()

	viper.Set("profile", *profile)
	viper.Set("configServerUrl", *configServerUrl)
	viper.Set("configBranch", *configBranch)
}

func main() {
	log.Printf("Starting %v\n", appName)

	config.LoadConfigurationFromBranch(
		viper.GetString("configServerUrl"),
		appName,
		viper.GetString("profile"),
		viper.GetString("configBranch"),
	)

	initializeBoltClient()
	initializeMessaging()
	initializeTracing()

	cb.ConfigureHystrix([]string{"imageservice", "quotes-service"}, service.MessagingClient)

	handleSigterm(func() {
		cb.Deregister(service.MessagingClient)
		service.MessagingClient.Close()
	})

	service.StartWebServer(viper.GetString("server_port"))
}

func initializeTracing() {
	tracing.InitTracing(viper.GetString("zipkin_server_url"), appName)
}

func handleSigterm(handleExit func()) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		<-c
		handleExit()
		os.Exit(1)
	}()
}
