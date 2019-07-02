package circuitbreaker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/linhnh123/golang-microservices-tutorial/common/util"

	"github.com/eapache/go-resiliency/retrier"
	"github.com/linhnh123/golang-microservices-tutorial/common/messaging"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/afex/hystrix-go/hystrix"
)

type DiscoveryToken struct {
	State   string `json:"state"` // UP, RUNNING, DOWN ??
	Address string `json:"address"`
}

func init() {
	log.SetOutput(ioutil.Discard)
}

var Client http.Client
var RETRIES = 3

func CallUsingCircuitBreaker(breakerName string, url string, method string) ([]byte, error) {
	output := make(chan []byte, 1)
	errors := hystrix.Go(breakerName, func() error {
		req, _ := http.NewRequest(method, url, nil)
		err := callWithRetriesreq(req, output)

		return err
	}, func(err error) error {
		log.Printf("In fallback function for breaker %v, error: %v", breakerName, err.Error())
		circuit, _, _ := hystrix.GetCircuit(breakerName)
		log.Printf("Circuit state is: %v", circuit.IsOpen())
		return err
	})

	select {
	case out := <-output:
		logrus.Debugf("Call in breaker %v successful", breakerName)
		return out, nil
	case err := <-errors:
		log.Printf("Error on channel in breaker %v. Message: %v", breakerName, err.Error())
		return nil, err
	}
}

func callWithRetriesreq(req *http.Request, output chan []byte) error {
	r := retrier.New(retrier.ConstantBackoff(RETRIES, 100*time.Millisecond), nil)
	attempt := 0
	err := r.Run(func() error {
		attempt++
		resp, err := Client.Do(req)
		if err == nil && resp.StatusCode < 299 {
			responseBody, err := ioutil.ReadAll(resp.Body)
			if err == nil {
				output <- responseBody
				return nil
			}
			return err
		} else if err == nil {
			err = fmt.Errorf("Status was %v", resp.StatusCode)
		}
		logrus.Errorf("Retrier failed, attempt %v", attempt)
		return err
	})
	return err
}

func ConfigureHystrix(commands []string, amqpClient messaging.IMessagingClient) {
	for _, command := range commands {
		hystrix.ConfigureCommand(command, hystrix.CommandConfig{
			Timeout:                resolveProperty(command, "Timeout"),
			MaxConcurrentRequests:  resolveProperty(command, "MaxConcurrentRequests"),
			ErrorPercentThreshold:  resolveProperty(command, "ErrorPercentThreshold"),
			RequestVolumeThreshold: resolveProperty(command, "RequestVolumeThreshold"),
			SleepWindow:            resolveProperty(command, "SleepWindow"),
		})
		logrus.Printf("Circuit %v settings: %v", command, hystrix.GetCircuitSettings()[command])
	}

	hystrixStreamHandler := hystrix.NewStreamHandler()
	hystrixStreamHandler.Start()
	go http.ListenAndServe(net.JoinHostPort("", "8181"), hystrixStreamHandler)
	logrus.Infoln("Launched hystrixStreamHandler at 8181")

	// Publish presence on RabbitMQ
	publishDiscoveryToken(amqpClient)
}

func resolveProperty(command string, prop string) int {
	if viper.IsSet("hystrix.command." + command + "." + prop) {
		return viper.GetInt("hystrix.command." + command + "." + prop)
	} else {
		return getDefaultHystrixConfigPropertyValue(prop)
	}
}

func publishDiscoveryToken(amqpClient messaging.IMessagingClient) {
	ip, err := util.ResolveIpFromHostsFile()
	if err != nil {
		ip = util.GetIPWithPrefix("10.0.")
	}
	token := DiscoveryToken{
		State:   "UP",
		Address: ip,
	}
	bytes, _ := json.Marshal(token)
	go func() {
		for {
			amqpClient.PublishOnQueue(bytes, "discovery")
			amqpClient.PublishOnQueue(bytes, "discovery")
			time.Sleep(time.Second * 30)
		}
	}()
}

func getDefaultHystrixConfigPropertyValue(prop string) int {
	switch prop {
	case "Timeout":
		return hystrix.DefaultTimeout
	case "MaxConcurrentRequests":
		return hystrix.DefaultMaxConcurrent
	case "RequestVolumeThreshold":
		return hystrix.DefaultVolumeThreshold
	case "SleepWindow":
		return hystrix.DefaultSleepWindow
	case "ErrorPercentThreshold":
		return hystrix.DefaultErrorPercentThreshold
	}
	panic("Got unknown hystrix property: " + prop + ". Panicing!")
}

func Deregister(amqpClient messaging.IMessagingClient) {
	ip, err := util.ResolveIpFromHostsFile()
	if err != nil {
		ip = util.GetIPWithPrefix("10.0.")
	}
	token := DiscoveryToken{
		State:   "DOWN",
		Address: ip,
	}
	bytes, _ := json.Marshal(token)
	amqpClient.PublishOnQueue(bytes, "discovery")
}
