package main

import (
	"contrib.go.opencensus.io/exporter/prometheus"
	"fmt"
	tracingexporter "github.com/tsocial/tracing/exporter"
	tracinggorm "github.com/tsocial/tracing/gorm"
	tracinghttp "github.com/tsocial/tracing/http"
	tracingredis "github.com/tsocial/tracing/redis"
	"github.com/tsocial/vite"
	"github.com/tsocial/vite/httpkit"
	"go.opencensus.io/trace"
	"log"
	"net/http"
	"os"
)

func main() {
	// important for testing
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
	address := "localhost:3000"
	if len(os.Args) > 1 {
		address = os.Args[1]
	}

	loadRedisConfig()
	MigrateDB()

	_, err := tracingexporter.RunJaegerExporter(
		fmt.Sprintf("trusting_social_demo||%s", address),
		"localhost:6831",
		"http://localhost:14268/api/traces",
	)
	if err != nil {
		panic(err)
	}

	err = tracinggorm.RegisterAllDatabaseViews()
	if err != nil {
		panic(err)
	}
	defer tracinggorm.UnregisterAllDatabaseViews()

	err = tracingredis.RegisterAllViews()
	if err != nil {
		panic(err)
	}
	defer tracingredis.UnregisterAllViews()

	//_, err = tracing.RunConsoleExporter()
	//if err != nil {
	//	panic(err)
	//}

	pe, err := tracingexporter.RunPrometheusExporter("trustingsocial_ocmetrics")
	if err != nil {
		panic(err)
	}

	// startRawHTTPServer(address, pe)

	err = startServerUsingHttpKit(address, pe)
	if err != nil {
		panic(err)
	}
}

func startRawHTTPServer(address string, pe *prometheus.Exporter) {
	mux := http.NewServeMux()
	mux.HandleFunc("/first", firstAPI)
	mux.HandleFunc("/second", secondAPI)

	// add endpoint for zpages and prometheus
	tracinghttp.AddLocalTraceEndpoint(mux, "/debug")
	tracinghttp.AddPrometheusEndpoint(mux, pe, "/metrics")

	handler, err := tracinghttp.WrapHandlerWithTracing(mux, OptionTracing)
	if err != nil {
		panic(err)
	}

	// start
	if err := http.ListenAndServe(address, handler); err != nil {
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
		TracingOption: &OptionTracing,
	}
	server := app.NewHTTPServer(address, option)

	// add local tracing with zpages
	// zpages.Handle(app.Mux, "/debug")

	// add api endpoint for prometheus
	// app.Mux.Handle("/metrics", pe)

	log.Println(vite.MarkInfo, "starting server", address)
	return server.Start()
}
