package service

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/linhnh123/golang-microservices-tutorial/common/messaging"

	"github.com/linhnh123/golang-microservices-tutorial/accountservice/model"

	"github.com/gorilla/mux"

	"github.com/linhnh123/golang-microservices-tutorial/accountservice/dbclient"
	cb "github.com/linhnh123/golang-microservices-tutorial/common/circuitbreaker"
)

var DBClient dbclient.IBoltClient
var MessagingClient messaging.IMessagingClient

type healthCheckResponse struct {
	Status string `json:"status"`
}

var client = &http.Client{}
var fallbackQuote = model.Quote{
	Language: "en",
	ServedBy: "circuit-breaker",
	Text:     "Text Breaker",
}

func init() {
	var transport http.RoundTripper = &http.Transport{
		DisableKeepAlives: true,
	}
	client.Transport = transport
}

func getIp() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "error"
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	panic("Unable to determine local IP address")
}

func notifyVIP(account model.Account) {
	if account.Id == "10000" {
		go func(account model.Account) {
			vipNotification := model.VipNotification{AccountId: account.Id, ReadAt: time.Now().UTC().String()}
			data, _ := json.Marshal(vipNotification)
			err := MessagingClient.PublishOnQueue(data, "vipQueue")
			if err != nil {
				log.Println(err.Error())
			}
		}(account)
	}
}

func GetAccount(w http.ResponseWriter, r *http.Request) {
	var accountId = mux.Vars(r)["accountId"]

	account, err := DBClient.QueryAccount(accountId)
	account.ServedBy = getIp()

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	notifyVIP(account)

	account.Quote = getQuote()
	account.ImageUrl = getImageUrl(accountId)

	data, _ := json.Marshal(account)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func getQuote() model.Quote {
	body, err := cb.CallUsingCircuitBreaker("quotes-service", "http://quotes-service:8080/api/quote?strength=4", "GET")
	if err == nil {
		quote := model.Quote{}
		json.Unmarshal(body, &quote)
		return quote
	} else {
		return fallbackQuote
	}
}

func writeJsonResponse(w http.ResponseWriter, status int, data []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(status)
	w.Write(data)
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	dbUp := DBClient.Check()
	if dbUp {
		data, _ := json.Marshal(healthCheckResponse{Status: "UP"})
		writeJsonResponse(w, http.StatusOK, data)
	} else {
		data, _ := json.Marshal(healthCheckResponse{Status: "Database unaccessible"})
		writeJsonResponse(w, http.StatusServiceUnavailable, data)
	}
}

func getImageUrl(accountId string) string {
	body, err := cb.CallUsingCircuitBreaker("imageservice", "http://imageservice:7777/accounts/"+accountId, "GET")
	if err == nil {
		return string(body)
	} else {
		return "http://path.to.placeholder"
	}
}
