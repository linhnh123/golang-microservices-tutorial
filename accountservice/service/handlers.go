package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/linhnh123/golang-microservices-tutorial/common/messaging"
	"github.com/linhnh123/golang-microservices-tutorial/common/util"
	"github.com/sirupsen/logrus"

	internalmodel "github.com/linhnh123/golang-microservices-tutorial/accountservice/model"

	"github.com/gorilla/mux"

	"github.com/linhnh123/golang-microservices-tutorial/accountservice/dbclient"
	cb "github.com/linhnh123/golang-microservices-tutorial/common/circuitbreaker"

	"github.com/linhnh123/golang-microservices-tutorial/common/model"
	"github.com/linhnh123/golang-microservices-tutorial/common/tracing"
)

var DBClient dbclient.IBoltClient
var MessagingClient messaging.IMessagingClient
var myIP string

type healthCheckResponse struct {
	Status string `json:"status"`
}

var client = &http.Client{}
var fallbackQuote = internalmodel.Quote{
	Language: "en",
	ServedBy: "circuit-breaker",
	Text:     "Text Breaker",
}

func init() {
	var transport http.RoundTripper = &http.Transport{
		DisableKeepAlives: true,
	}
	client.Transport = transport
	cb.Client = *client
	var err error
	myIP, err = util.ResolveIpFromHostsFile()
	if err != nil {
		myIP = util.GetIp()
	}
	fmt.Println("Init method executed")
}

func notifyVIP(ctx context.Context, account internalmodel.Account) {
	if account.Id == "10000" {
		go func(account internalmodel.Account) {
			vipNotification := internalmodel.VipNotification{AccountId: account.Id, ReadAt: time.Now().UTC().String()}
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

	account, err := DBClient.QueryAccount(r.Context(), accountId)
	account.ServedBy = util.GetIp()

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	notifyVIP(r.Context(), account)

	account.Quote = getQuote(r.Context())
	account.ImageData = getImageUrl(r.Context(), accountId)

	data, _ := json.Marshal(account)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func getQuote(ctx context.Context) internalmodel.Quote {
	body, err := cb.CallUsingCircuitBreaker("quotes-service", "http://quotes-service:8080/api/quote?strength=4", "GET")
	if err == nil {
		quote := internalmodel.Quote{}
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

func getImageUrl(ctx context.Context, accountId string) model.AccountImage {
	child := tracing.StartSpanFromContextWithLogEvent(ctx, "getImageUrl", "Client send")
	defer tracing.CloseSpan(child, "Client Receive")

	req, err := http.NewRequest("GET", "http://imageservice:7777/accounts/"+accountId, nil)
	body, err := cb.PerformHTTPRequestCircuitBreaker(tracing.UpdateContext(ctx, child), "account-to-image", req)
	if err == nil {
		accountImage := model.AccountImage{}
		err := json.Unmarshal(body, &accountImage)
		if err == nil {
			return accountImage
		}
		panic("Unmarshalling accountImage struct went really bad. Msg: " + err.Error())
	}
	return model.AccountImage{URL: "http://path.to.placeholder", ServedBy: "fallback"}

}

func fetchAccount(ctx context.Context, accountID string) (internalmodel.Account, error) {
	account, err := getAccount(ctx, accountID)
	if err != nil {
		return account, err
	}
	account.Quote = getQuote(ctx)
	account.ImageData = getImageUrl(ctx, accountID)
	account.ServedBy = myIP

	notifyVIP(ctx, account) // Send VIP notification concurrently.

	// If found, marshal into JSON, write headers and content
	return account, nil
}

func getAccount(ctx context.Context, accountID string) (internalmodel.Account, error) {
	// Start a new opentracing child span
	child := tracing.StartSpanFromContextWithLogEvent(ctx, "getAccountData", "Client send")
	defer tracing.CloseSpan(child, "Client Receive")

	// Create the http request and pass it to the circuit breaker
	req, err := http.NewRequest("GET", "http://dataservice:7070/accounts/"+accountID, nil)
	body, err := cb.PerformHTTPRequestCircuitBreaker(tracing.UpdateContext(ctx, child), "account-to-data", req)
	if err == nil {
		accountData := model.AccountData{}
		json.Unmarshal(body, &accountData)
		return toAccount(accountData), nil
	}
	logrus.Errorf("Error: %v\n", err.Error())
	return internalmodel.Account{}, err
}

func toAccount(accountData model.AccountData) internalmodel.Account {
	return internalmodel.Account{
		Id: accountData.ID, Name: accountData.Name, AccountEvents: accountData.Events,
	}
}
