package main

import (
	"context"
	"contrib.go.opencensus.io/exporter/prometheus"
	"crypto/tls"
	"github.com/tsocial/vite"
	"github.com/tsocial/vite/httpkit"
	"github.com/tsocial/vite/httpkit/comm"
	"github.com/tsocial/vite/tracing"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/plugin/ochttp/propagation/tracecontext"
	"go.opencensus.io/trace"
	"go.opencensus.io/zpages"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var requestOption = comm.CommRequestOption{
	Transport: &ochttp.Transport{
		Propagation: &tracecontext.HTTPFormat{},
		// Propagation: &b3.HTTPFormat{},
		Base: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	},
}

func secondAPI(w http.ResponseWriter, r *http.Request) {
	log.Println(vite.MarkError, "thao")
	log.Println(vite.MarkError, r.Header)
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
	// orm = tracing.GormWithContext(r.Context(), testDB)
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

func main() {
	address := "localhost:3000"
	if len(os.Args) > 1 {
		address = os.Args[1]
	}

	loadRedisConfig()
	MigrateDB()

	_, err := tracing.RunJaegerExporter(
		"trusting_social_demo",
		"localhost:6831",
		"http://localhost:14268/api/traces",
	)
	if err != nil {
		panic(err)
	}

	err = tracing.RegisterAllDatabaseViews()
	if err != nil {
		panic(err)
	}
	defer tracing.UnregisterAllDatabaseViews()

	err = tracing.RegisterAllRedisViews()
	if err != nil {
		panic(err)
	}
	defer tracing.UnregisterAllRedisViews()

	_, err = tracing.RunConsoleExporter()
	if err != nil {
		panic(err)
	}

	pe, err := tracing.RunPrometheusExporter("trustingsocial_ocmetrics")
	if err != nil {
		panic(err)
	}
	// startRawHTTPServer(pe)

	err = startServerUsingHttpKit(address, pe)
	if err != nil {
		panic(err)
	}
}

func startRawHTTPServer(address string, pe *prometheus.Exporter) {
	mux := http.NewServeMux()
	mux.HandleFunc("/first", firstAPI)

	// add local tracing with zpages
	zpages.Handle(mux, "/debug")

	// add api endpoint for prometheus
	mux.Handle("/metrics", pe)

	// wrap handler inside OpenCensus handler for tracing request
	och := &ochttp.Handler{
		Handler: mux,
	}

	// start
	if err := http.ListenAndServe(address, och); err != nil {
		panic(err)
	}
}

func startServerUsingHttpKit(address string, pe *prometheus.Exporter) error {
	// create app object
	handlers := []*httpkit.RouteHandler{
		{
			Route: &httpkit.Route{
				Name:   "first_api",
				Method: http.MethodGet,
				Path:   "/first",
			},
			Handle: firstAPI,
		},
		{
			Route: &httpkit.Route{
				Name:   "second_api",
				Method: http.MethodGet,
				Path:   "/second",
			},
			Handle: secondAPI,
		},
	}
	app := httpkit.NewApp(nil, httpkit.SampleSecret)
	app.AddPublicRouteHandlers(handlers...)

	// create server object
	option := httpkit.ServerOption{
		TracingOption: &tracing.OptionTracing{
			PropagationFormat: tracing.TracingContextFormat,
			IsPublicEndpoint:  false,
			SamplingFraction:  1,
		},
	}
	server := app.NewHTTPServer(address, option)

	// decorated server object
	// TODO looking for better way

	// add local tracing with zpages
	zpages.Handle(app.Mux, "/debug")

	// add api endpoint for prometheus
	app.Mux.Handle("/metrics", pe)

	log.Println(vite.MarkInfo, "starting server", address)
	return server.Start()
}
