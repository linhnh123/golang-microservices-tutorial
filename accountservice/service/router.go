package service

import (
	"net/http"

	"github.com/linhnh123/golang-microservices-tutorial/common/tracing"

	"github.com/gorilla/mux"
)

func NewRouter() *mux.Router {
	initQL(&LiveGraphQLResolvers{})

	router := mux.NewRouter().StrictSlash(true)

	for _, route := range routes {
		router.Methods(route.Method).Path(route.Pattern).Name(route.Name).Handler(loadTracing(route.HandlerFunc))
	}

	return router
}

func loadTracing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		span := tracing.StartHTTPTrace(req, "GetAccount") // Start the span
		ctx := tracing.UpdateContext(req.Context(), span) // Add span to context
		next.ServeHTTP(rw, req.WithContext(ctx))          // Note next-based chaining and copy of context!!
		span.Finish()                                     // Finish the span
	})
}
