package main

import (
	"context"
	tracinggorm "github.com/tsocial/tracing/gorm"
	tracinghttp "github.com/tsocial/tracing/http"
	tracingredis "github.com/tsocial/tracing/redis"
	"github.com/tsocial/vite"
	"github.com/tsocial/vite/httpkit"
	"github.com/tsocial/vite/httpkit/comm"
	"go.opencensus.io/trace"
	"log"
	"net/http"
	"strings"
	"time"
)

var OptionTracing = tracinghttp.OptionTracing{
	PropagationFormat: tracinghttp.B3Format,
	IsPublicEndpoint:  false,
	SamplingFraction:  1.0,
	ServiceName:       "distributed_tracing_demo",
}

var TracingTransport, _ = tracinghttp.WrapTransportWithTracing(httpkit.DefaultTransport, OptionTracing)

var CommRequestOption = &comm.RequestOption{
	Transport: TracingTransport,
}

func secondAPI(w http.ResponseWriter, r *http.Request) {
	_, span := trace.StartSpan(r.Context(), "second api span")
	span.AddAttributes(trace.StringAttribute("api", "second api"))
	time.Sleep(1 * time.Second)
	span.End()

	_, _ = w.Write([]byte("Hello! I am second API"))
}

func firstAPI(w http.ResponseWriter, r *http.Request) {
	message := r.URL.Path
	message = strings.TrimPrefix(message, "/")
	message = "Hello " + message

	// wrap orm object before calling database
	orm := tracinggorm.WithContext(r.Context(), testDB)
	product, _ := GetFirstProductWithContext(orm)
	log.Println(product.Name)

	// different type of query
	_ = GetAllProductDates(orm)

	// wrap redis object before calling redis operator
	wrapRedis := tracingredis.WithContext(r.Context(), Redis.Client)
	writeKeyWithContext(wrapRedis, "service", "StackOverFlow")
	val := readKeyWithContext(wrapRedis, "service")
	log.Println(val)

	// unfound key. expected tracing backend will show error
	val = readKeyWithContext(wrapRedis, "UNFOUND_KEY")
	time.Sleep(1 * time.Second)

	// manually create span
	_, span := trace.StartSpan(r.Context(), "child")
	span.Annotate([]trace.Attribute{trace.StringAttribute("key", "value")}, "something happened")
	span.AddAttributes(trace.StringAttribute("hello", "world"))
	time.Sleep(time.Millisecond * 125)
	span.End()

	// call external API.
	sendExternalRequest(r.Context())
	sendExternalRequest(r.Context())

	// call internal API
	sendInternalRequest(r.Context())

	// manually create span
	_, span = trace.StartSpan(r.Context(), "second child")
	span.Annotate([]trace.Attribute{trace.StringAttribute("key", "value")}, "something happened")
	span.AddAttributes(trace.StringAttribute("ackack", "ack ack"))
	time.Sleep(time.Millisecond * 125)
	span.End()

	_, _ = w.Write([]byte(message))
}

func sendExternalRequest(ctx context.Context) {
	url := "https://example.com"
	request := comm.NewRequest(ctx, http.MethodGet, url, vite.Map{}, nil, CommRequestOption)
	outData := vite.Map{}
	_, _ = request.Send(&outData)
}

func sendInternalRequest(ctx context.Context) {
	url := "http://localhost:4000/second"
	request := comm.NewRequest(ctx, http.MethodGet, url, vite.Map{}, nil, CommRequestOption)
	outData := vite.Map{}
	_, _ = request.Send(&outData)
}
