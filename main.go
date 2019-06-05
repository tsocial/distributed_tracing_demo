package main

import (
	"contrib.go.opencensus.io/exporter/prometheus"
	"github.com/tsocial/vite"
	"github.com/tsocial/vite/httpkit"
	"github.com/tsocial/vite/tracing"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/trace"
	"go.opencensus.io/zpages"
	"log"
	"net/http"
	"os"
)

func main() {
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
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

	//_, err = tracing.RunConsoleExporter()
	//if err != nil {
	//	panic(err)
	//}

	pe, err := tracing.RunPrometheusExporter("trustingsocial_ocmetrics")
	if err != nil {
		panic(err)
	}
	startRawHTTPServer(address, pe)

	//err = startServerUsingHttpKit(address, pe)
	//if err != nil {
	//	panic(err)
	//}
}

func startRawHTTPServer(address string, pe *prometheus.Exporter) {
	mux := http.NewServeMux()
	mux.HandleFunc("/first", firstAPI)
	mux.HandleFunc("/second", secondAPI)

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
