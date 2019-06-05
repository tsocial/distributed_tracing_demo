package main

import (
	"context"
	"crypto/tls"
	"github.com/tsocial/vite"
	"github.com/tsocial/vite/httpkit/comm"
	"github.com/tsocial/vite/tracing"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/plugin/ochttp/propagation/tracecontext"
	"go.opencensus.io/trace"
	"log"
	"net/http"
	"strings"
	"time"
)

var requestOption = comm.RequestOption{
	Transport: &ochttp.Transport{
		Propagation: &tracecontext.HTTPFormat{},
		// Propagation: &b3.HTTPFormat{},
		Base: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	},
}

func secondAPI(w http.ResponseWriter, r *http.Request) {
	time.Sleep(1 * time.Second)
	_, _ = w.Write([]byte("Hello! I am second API"))
}

func firstAPI(w http.ResponseWriter, r *http.Request) {
	message := r.URL.Path
	message = strings.TrimPrefix(message, "/")
	message = "Hello " + message

	// wrap orm object before calling database
	orm := tracing.GormWithContext(r.Context(), testDB)
	product, _ := GetFirstProductWithContext(orm)
	log.Println(product.Name)

	// different type of query
	_ = GetAllProductDates(orm)

	// wrap redis object before calling redis operator
	wrapRedis := tracing.RedisWithContext(r.Context(), Redis.Client)
	writeKeyWithContext(wrapRedis, "service", "StackOverFlow")
	val := readKeyWithContext(wrapRedis, "service")
	log.Println(val)

	// unfound key. expected tracing backend will show error
	val = readKeyWithContext(wrapRedis, "UNFOUND_KEY")
	time.Sleep(1 * time.Second)

	// manually create span
	_, span := trace.StartSpan(r.Context(), "child")
	defer span.End()
	span.Annotate([]trace.Attribute{trace.StringAttribute("key", "value")}, "something happened")
	span.AddAttributes(trace.StringAttribute("hello", "world"))
	time.Sleep(time.Millisecond * 125)

	// call external API.
	sendExternalRequest(r.Context())

	// call internal API
	sendInternalRequest(r.Context())

	// manually create span
	_, span = trace.StartSpan(r.Context(), "child")
	defer span.End()
	span.Annotate([]trace.Attribute{trace.StringAttribute("key", "value")}, "something happened")
	span.AddAttributes(trace.StringAttribute("hello", "world"))
	time.Sleep(time.Millisecond * 125)

	_, _ = w.Write([]byte(message))
}

func sendExternalRequest(ctx context.Context) {
	url := "https://example.com"
	request := comm.NewRequestWithContext(ctx, http.MethodGet, url, vite.Map{}, nil, requestOption)
	outData := vite.Map{}
	_, _ = request.Send(&outData)
}

func sendInternalRequest(ctx context.Context) {
	url := "http://localhost:4000/second"
	request := comm.NewRequestWithContext(ctx, http.MethodGet, url, vite.Map{}, nil, requestOption)
	outData := vite.Map{}
	_, _ = request.Send(&outData)
}
