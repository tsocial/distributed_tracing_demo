package main

import (
	"github.com/tsocial/vite/tracing"
	"go.opencensus.io/trace"
	"go.opencensus.io/zpages"
	"log"
	"net/http"
	"strings"
	"time"

	"contrib.go.opencensus.io/exporter/prometheus"
	"go.opencensus.io/plugin/ochttp"
)

func loginAPI(w http.ResponseWriter, r *http.Request) {
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

	// manually create span
	_, span := trace.StartSpan(r.Context(), "child")
	defer span.End()
	span.Annotate([]trace.Attribute{trace.StringAttribute("key", "value")}, "something happened")
	span.AddAttributes(trace.StringAttribute("hello", "world"))
	time.Sleep(time.Millisecond * 125)

	// call external API.
	nr, _ := http.NewRequest("GET", "https://example.com", nil)
	// Propagate the trace header info in the outgoing requests.
	nr = nr.WithContext(r.Context())
	client := &http.Client{Transport: &ochttp.Transport{}}
	response, err := client.Do(nr)
	if err != nil {
		log.Println(err)
	}
	_ = response.Body.Close()

	_, _ = w.Write([]byte(message))
}

func main() {
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

	_, err = tracing.RunConsoleExporter()
	if err != nil {
		panic(err)
	}

	initServer(nil)
}

func initServer(pe *prometheus.Exporter) {
	mux := http.NewServeMux()
	mux.HandleFunc("/login", loginAPI)

	// add local tracing with zpages
	zpages.Handle(mux, "/debug")

	// add api endpoint for prometheus
	// mux.Handle("/metrics", pe)

	och := &ochttp.Handler{
		Handler: mux,
	}

	if err := http.ListenAndServe(":3000", och); err != nil {
		panic(err)
	}
}
