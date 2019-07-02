package tracing

import (
	"context"
	"fmt"
	"net/http"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/sirupsen/logrus"

	zipkin "github.com/openzipkin/zipkin-go-opentracing"
)

var tracer opentracing.Tracer

// SetTracer can be used by unit tests to provide a NoopTracer instance. Real users should always
// use the InitTracing func.
func SetTracer(initializedTracer opentracing.Tracer) {
	tracer = initializedTracer
}

// InitTracing connects the calling service to Zipkin and initializes the tracer.
func InitTracing(zipkinURL string, serviceName string) {
	logrus.Infof("Connecting to zipkin server at %v", zipkinURL)
	collector, err := zipkin.NewHTTPCollector(
		fmt.Sprintf("%s/api/v1/spans", zipkinURL))
	if err != nil {
		logrus.Info("Error connecting to zipkin server at " +
			fmt.Sprintf("%s/api/v1/spans", zipkinURL) + ". Error: " + err.Error())
		logrus.Errorln("Error connecting to zipkin server at " +
			fmt.Sprintf("%s/api/v1/spans", zipkinURL) + ". Error: " + err.Error())
		panic("Error connecting to zipkin server at " +
			fmt.Sprintf("%s/api/v1/spans", zipkinURL) + ". Error: " + err.Error())
	}
	tracer, err = zipkin.NewTracer(
		zipkin.NewRecorder(collector, false, "127.0.0.1:0", serviceName))
	if err != nil {
		logrus.Errorln("Error starting new zipkin tracer. Error: " + err.Error())
		panic("Error starting new zipkin tracer. Error: " + err.Error())
	}
	logrus.Infof("Successfully started zipkin tracer for service '%v'", serviceName)
}

// StartHTTPTrace loads tracing information from an INCOMING HTTP request.
func StartHTTPTrace(r *http.Request, opName string) opentracing.Span {
	carrier := opentracing.HTTPHeadersCarrier(r.Header)
	clientContext, err := tracer.Extract(opentracing.HTTPHeaders, carrier)
	if err == nil {
		return tracer.StartSpan(
			opName, ext.RPCServerOption(clientContext))
	} else {
		return tracer.StartSpan(opName)
	}
}

// UpdateContext updates the supplied context with the supplied span.
func UpdateContext(ctx context.Context, span opentracing.Span) context.Context {
	return context.WithValue(ctx, "opentracing-span", span)
}

// StartChildSpanFromContext starts a child span from span within the supplied context, if available.
func StartChildSpanFromContext(ctx context.Context, opName string) opentracing.Span {
	if ctx.Value("opentracing-span") == nil {
		return tracer.StartSpan(opName, ext.RPCServerOption(nil))
	}
	parent := ctx.Value("opentracing-span").(opentracing.Span)
	child := tracer.StartSpan(opName, opentracing.ChildOf(parent.Context()))
	return child
}

func StartSpanFromContextWithLogEvent(ctx context.Context, opName string, logStatement string) opentracing.Span {
	span := ctx.Value("opentracing-span").(opentracing.Span)
	child := tracer.StartSpan(opName, ext.RPCServerOption(span.Context()))
	child.LogEvent(logStatement)
	return child
}

func CloseSpan(span opentracing.Span, event string) {
	span.LogEvent(event)
	span.Finish()
}

func AddTracingToReqFromContext(ctx context.Context, req *http.Request) {
	if ctx.Value("opentracing-span") == nil {
		return
	}
	carrier := opentracing.HTTPHeadersCarrier(req.Header)
	err := tracer.Inject(
		ctx.Value("opentracing-span").(opentracing.Span).Context(),
		opentracing.HTTPHeaders,
		carrier)
	if err != nil {
		panic("Unable to inject tracing context: " + err.Error())
	}
}
