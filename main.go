package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"contrib.go.opencensus.io/exporter/prometheus"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
	"gocloud.dev/server"
	"gocloud.dev/server/health"
)

type customHealthCheck struct {
	mu      sync.RWMutex
	healthy bool
}

func (h *customHealthCheck) CheckHealth() error {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if !h.healthy {
		return errors.New("not ready yet!")
	}
	return nil
}

func mainHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "Welcome to the home page!\n")
}

func helloHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "Hello\n")
}

type customTraceExporter struct{}

func (ce *customTraceExporter) ExportSpan(sd *trace.SpanData) {
	fmt.Printf("Name: %s\nTraceID: %x\nSpanID: %x\nParentSpanID: %x\nStartTime: %s\nEndTime: %s\nAnnotations: %+v\n\n",
		sd.Name, sd.TraceID, sd.SpanID, sd.ParentSpanID, sd.StartTime, sd.EndTime, sd.Annotations)
}

func main() {
	pe, err := prometheus.NewExporter(prometheus.Options{
		Namespace: "demo",
	})
	if err != nil {
		log.Fatalf("Failed to create Prometheus exporter: %v", err)
	}

	view.RegisterExporter(pe)

	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
	var exp trace.Exporter
	exp = &customTraceExporter{}
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", helloHandler)
	mux.Handle("/metrics", pe)
	mux.HandleFunc("/", mainHandler)

	// healthCheck will report the server is unhealthy for 10 seconds after
	// startup, and as healthy henceforth. Check the /healthz/readiness
	// HTTP path to see readiness.
	healthCheck := new(customHealthCheck)
	time.AfterFunc(10*time.Second, func() {
		healthCheck.mu.Lock()
		defer healthCheck.mu.Unlock()
		healthCheck.healthy = true
	})

	options := &server.Options{
		HealthChecks:  []health.Checker{healthCheck},
		TraceExporter: exp,

		// In production you will likely want to use trace.ProbabilitySampler
		// instead, since AlwaysSample will start and export a trace for every
		// request - this may be prohibitively slow with significant traffic.
		DefaultSamplingPolicy: trace.AlwaysSample(),
		Driver:                &server.DefaultDriver{},
	}

	s := server.New(mux, options)
	fmt.Printf("Listening on %s\n", ":8080")

	err = s.ListenAndServe(":8080")
	if err != nil {
		log.Fatal(err)
	}
}
