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
	"log"
	"net/http"
	"os"
)

func main() {
	// important for testing
	// don't need this config because have tracing option already
	// trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
	address := "localhost:3000"
	if len(os.Args) > 1 {
		address = os.Args[1]
	}

	loadRedisConfig()
	MigrateDB()

	pe := registerExporters(address)

	registerViews()
	defer unRegisterAllViews()

	// startRawHTTPServer(address, pe)

	err := startServerUsingHttpKit(address, pe)
	if err != nil {
		panic(err)
	}
}

// registerExporters register all possible exporters
func registerExporters(address string) *prometheus.Exporter {
	_, err := tracingexporter.RunJaegerExporter(
		fmt.Sprintf("trusting_social_demo||%s", address),
		"localhost:6831",
		"http://localhost:14268/api/traces",
	)
	if err != nil {
		panic(err)
	}

	//_, err = tracing.RunConsoleExporter()
	//if err != nil {
	//	panic(err)
	//}

	pe, err := tracingexporter.RunPrometheusExporter("trustingsocial_ocmetrics")
	if err != nil {
		panic(err)
	}
	return pe
}

// registerViews register all types of views framework can provided
func registerViews() {
	err := tracinggorm.RegisterAllDatabaseViews()
	if err != nil {
		panic(err)
	}

	err = tracingredis.RegisterAllViews()
	if err != nil {
		panic(err)
	}

	err = tracinghttp.RegisterAllViews()
	if err != nil {
		panic(err)
	}
}

func unRegisterAllViews() {
	tracinggorm.UnregisterAllDatabaseViews()
	tracingredis.UnregisterAllViews()
	tracinghttp.UnregisterAllViews()

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

	tracinghttp.AddLocalTraceEndpoint(app.Mux(), "/debug")
	tracinghttp.AddPrometheusEndpoint(app.Mux(), pe, "/metrics")

	log.Println(vite.MarkInfo, "starting server", address)
	return server.Start()
}
