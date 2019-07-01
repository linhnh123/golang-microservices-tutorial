package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"net"
	"os"
	"sync"

	"github.com/linhnh123/golang-microservices-tutorial/gelftail/aggregator"
	"github.com/linhnh123/golang-microservices-tutorial/gelftail/transformer"
	"github.com/sirupsen/logrus"
)

var authToken = ""
var port *string

func init() {
	data, err := ioutil.ReadFile("token.txt")
	if err != nil {
		msg := "Cannot find token.txt"
		logrus.Errorln(msg)
		panic(msg)
	}
	authToken = string(data)
	port = flag.String("port", "12202", "UDP port for gelftail")
	flag.Parse()
}

func main() {
	logrus.Println("Starting Gelf-tail server...")

	serverConn := startUDPServer(*port)
	defer serverConn.Close()

	var bulkQueue = make(chan []byte, 1)

	go aggregator.Start(bulkQueue, authToken)
	go listenForLogStatements(serverConn, bulkQueue)

	logrus.Infoln("Started Gelf-tail server")

	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}

func checkError(err error) {
	if err != nil {
		logrus.Println("Error: ", err)
		os.Exit(0)
	}
}

func startUDPServer(port string) *net.UDPConn {
	serverAddr, err := net.ResolveUDPAddr("udp", ":"+port)
	checkError(err)

	serverConn, err := net.ListenUDP("udp", serverAddr)
	checkError(err)

	return serverConn
}

func listenForLogStatements(serverConn *net.UDPConn, bulkQueue chan []byte) {
	buf := make([]byte, 8192) // 8kb
	var item map[string]interface{}

	for {
		n, _, err := serverConn.ReadFromUDP(buf)
		if err != nil {
			logrus.Errorf("Problem reading UDP message into buffer: %v\n", err.Error())
			continue
		}
		err = json.Unmarshal(buf[0:n], &item)
		if err != nil {
			logrus.Errorf("Problem unmarshalling log message into JSON: " + err.Error())
			item = nil
			continue
		}
		processedLogMessage, err := transformer.ProcessLogStatement(item)
		if err != nil {
			logrus.Printf("Problem parsing message: %v", string(buf[0:n]))
		} else {
			bulkQueue <- processedLogMessage
		}
		item = nil
	}
}
