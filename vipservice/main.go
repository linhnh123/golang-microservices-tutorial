package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/linhnh123/golang-microservices-tutorial/common/config"
	"github.com/linhnh123/golang-microservices-tutorial/vipservice/service"
	"github.com/streadway/amqp"

	"github.com/linhnh123/golang-microservices-tutorial/common/messaging"
	"github.com/spf13/viper"
)

var appName = "vipservice"

var messagingClient messaging.IMessagingClient

func init() {
	configServerUrl := flag.String("configServerUrl", "http://configserver:8888", "Address to config server")
	profile := flag.String("profile", "test", "Environment profile, something similar to spring profiles")
	configBranch := flag.String("configBranch", "master", "git branch to fetch configuration from")
	flag.Parse()

	viper.Set("profile", *profile)
	viper.Set("configServerUrl", *configServerUrl)
	viper.Set("configBranch", *configBranch)
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Printf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func onMessage(delivery amqp.Delivery) {
	log.Printf("Got a message: %v\n", string(delivery.Body))
}

func initializeMessaging() {
	if !viper.IsSet("amqp_server_url") {
		panic("No 'broker_url' set in configuration, cannot start")
	}
	messagingClient = &messaging.MessagingClient{}
	messagingClient.ConnectToBroker(viper.GetString("amqp_server_url"))

	// Call the subscribe method with queue name and callback function
	err := messagingClient.SubscribeToQueue("vipQueue", appName, onMessage)
	failOnError(err, "Could not start subscribe to vipQueue")

	err = messagingClient.Subscribe(viper.GetString("config_event_bus"), "topic", appName, config.HandleRefreshEvent)
	failOnError(err, "Could not start subscribe to "+viper.GetString("config_event_bus")+" topic")
}

func main() {
	log.Println("Starting " + appName + "...")

	config.LoadConfigurationFromBranch(viper.GetString("configServerUrl"), appName, viper.GetString("profile"), viper.GetString("configBranch"))

	initializeMessaging()

	handleSigterm(func() {
		if messagingClient != nil {
			messagingClient.Close()
		}
	})

	service.StartWebServer(viper.GetString("server_port"))
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
