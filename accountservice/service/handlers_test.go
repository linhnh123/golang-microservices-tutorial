package service

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/linhnh123/golang-microservices-tutorial/common/messaging"

	"github.com/stretchr/testify/mock"

	"gopkg.in/h2non/gock.v1"

	"github.com/linhnh123/golang-microservices-tutorial/accountservice/model"

	"github.com/linhnh123/golang-microservices-tutorial/accountservice/dbclient"
	. "github.com/smartystreets/goconvey/convey"
)

// Create a mock instance that implements the IBoltClient interface
var mockRepo = &dbclient.MockBoltClient{}
var mockMessagingClient = &messaging.MockMessagingClient{}

var anyString = mock.AnythingOfType("string")
var anyByteArray = mock.AnythingOfType("[]uint8")

func init() {
	gock.InterceptClient(client)
}

func TestGetAccountWrongPath(t *testing.T) {
	Convey("Given a HTTP request for /invalid/123", t, func() {
		req := httptest.NewRequest("GET", "/invalid/123", nil)
		resp := httptest.NewRecorder()

		Convey("When the request is handled by the Router", func() {
			NewRouter().ServeHTTP(resp, req)

			Convey("Then the response should be a 404", func() {
				So(resp.Code, ShouldEqual, 404)
			})
		})
	})
}

func TestGetAccount(t *testing.T) {
	defer gock.Off()
	gock.New("http://quotes-service:8080").
		Get("/api/quote").
		MatchParam("strength", "4").
		Reply(200).
		BodyString(`{"quote":"May the source be with you. Always","ipAddress":"10.0.0.5:8080","language":"en"}`)

	// Declare two mock behaviours. For "123" as input, return a proper Account struct and nil as error.
	// For "456" as input, return an empty Account object and a real error.
	mockRepo.On("QueryAccount", mock.Anything, "123").Return(model.Account{Id: "123", Name: "Person_123"}, nil)
	mockRepo.On("QueryAccount", mock.Anything, "456").Return(model.Account{}, fmt.Errorf("Some error"))

	// Finally, assign mockRepo to the DBClient field (it's in _handlers.go_, e.g. in the same package)
	DBClient = mockRepo

	Convey("Given a HTTP request for /accounts/123", t, func() {
		req := httptest.NewRequest("GET", "/accounts/123", nil)
		resp := httptest.NewRecorder()

		Convey("When the request is handled by the Router", func() {
			NewRouter().ServeHTTP(resp, req)

			Convey("Then the response should be a 200", func() {
				So(resp.Code, ShouldEqual, 200)

				account := model.Account{}
				json.Unmarshal(resp.Body.Bytes(), &account)
				So(account.Id, ShouldEqual, "123")
				So(account.Name, ShouldEqual, "Person_123")

				So(account.Quote.Text, ShouldEqual, "May the source be with you. Always")
			})
		})
	})

	Convey("Given a HTTP request for /accounts/456", t, func() {
		req := httptest.NewRequest("GET", "/accounts/456", nil)
		resp := httptest.NewRecorder()
		Convey("When the request is handled by the Router", func() {
			NewRouter().ServeHTTP(resp, req)
			Convey("Then the response should be a 404", func() {
				So(resp.Code, ShouldEqual, 404)
			})
		})
	})
}

func TestNotificationIsSentForVIPAccount(t *testing.T) {
	defer gock.Off()
	gock.New("http://quotes-service:8080").
		Get("/api/quote").
		MatchParam("strength", "4").
		Reply(200).
		BodyString(`{"quote":"May the source be with you. Always","ipAddress":"10.0.0.5:8080","language":"en"}`)

	mockRepo.On("QueryAccount", "10000").Return(model.Account{Id: "10000", Name: "Person_123"}, nil)
	DBClient = mockRepo

	mockMessagingClient.On("PublishOnQueue", anyByteArray, anyString).Return(nil)
	MessagingClient = mockMessagingClient

	Convey("Given a HTTP req for a VIP account", t, func() {
		req := httptest.NewRequest("GET", "/accounts/10000", nil)
		resp := httptest.NewRecorder()

		Convey("When the request is handled by the Router", func() {
			NewRouter().ServeHTTP(resp, req)
			Convey("Then the response should be a 200 and the MessageClient should have been invoked", func() {
				So(resp.Code, ShouldEqual, 200)
				time.Sleep(time.Millisecond * 100)
				So(mockMessagingClient.AssertNumberOfCalls(t, "PublishOnQueue", 1), ShouldBeTrue)
			})
		})
	})
}
