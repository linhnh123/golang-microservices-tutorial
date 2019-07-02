package service

import (
	"net/http"

	gqlhandler "github.com/graphql-go/graphql-go-handler"
)

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

var routes = Routes{
	Route{
		"GetAccount",
		"GET",
		"/accounts/{accountId}",
		GetAccount,
	},
	Route{
		"HealthCheck",
		"GET",
		"/health",
		HealthCheck,
	},
	Route{
		"GraphQL",
		"POST",
		"/graphql",
		gqlhandler.New(&gqlhandler.Config{
			Schema: &schema,
			Pretty: false,
		}).ServeHTTP,
	},
}
