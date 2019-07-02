package main

import (
	"flag"
	"sync"
	"time"

	"github.com/linhnh123/golang-microservices-tutorial/common/config"
	"github.com/linhnh123/golang-microservices-tutorial/common/messaging"
	"github.com/linhnh123/golang-microservices-tutorial/imageservice/service"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var appName = "imageservice"

func init() {
	profile := flag.String("profile", "test", "Environment profile, something similar to spring profiles")
	configServerUrl := flag.String("configServerUrl", "http://configserver:8888", "Address to config server")
	configBranch := flag.String("configBranch", "master", "git branch to fetch configuration from")

	flag.Parse()

	viper.Set("profile", *profile)
	viper.Set("configServerUrl", *configServerUrl)
	viper.Set("configBranch", *configBranch)
}

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.Infof("Starting %v", appName)

	start := time.Now().UTC()
	config.LoadConfigurationFromBranch(viper.GetString("configServerUrl"), appName, viper.GetString("profile"), viper.GetString("configBranch"))
	initializeMessaging()
	go service.StartWebServer(viper.GetString("server_port")) // Starts HTTP service  (async)

	logrus.Infof("Started %v in %v", appName, time.Now().UTC().Sub(start))
	// Block...
	wg := sync.WaitGroup{} // Use a WaitGroup to block main() exit
	wg.Add(1)
	wg.Wait()
}

func initializeMessaging() {
	if !viper.IsSet("amqp_server_url") {
		panic("No 'amqp_server_url' set in configuration, cannot start")
	}

	service.MessagingClient = &messaging.MessagingClient{}
	service.MessagingClient.ConnectToBroker(viper.GetString("amqp_server_url"))
	service.MessagingClient.Subscribe(viper.GetString("config_event_bus"), "topic", appName, config.HandleRefreshEvent)
}
