package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func fastHandler(w http.ResponseWriter, r *http.Request) {
	totalRequests.Inc()
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		requestDuration.WithLabelValues("fast").Observe(duration.Seconds())
	}()
	fmt.Fprintf(w, "This is a fast API endpoint")
}

func slowHandler(w http.ResponseWriter, r *http.Request) {
	totalRequests.Inc()
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		requestDuration.WithLabelValues("slow").Observe(duration.Seconds())
	}()
	time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)
	if rand.Intn(10) < 2 {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"status": "error", "message": "Internal Server Error"}`)
		return
	}

	fmt.Fprintf(w, `{"status": "success", "message": "Slow API call processed"}`)
}

var (
	totalRequests = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "total_requests_count",
			Help: "Total number of API requests",
		},
	)
	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "api_request_duration_seconds",
			Help:    "Distribution of request duration in seconds",
			Buckets: prometheus.LinearBuckets(0.001, 0.01, 10), // Adjust bucket range as needed
		},
		[]string{"endpoint"},
	)
	counter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_requests_total",
			Help: "Total number of API requests",
		},
		[]string{"endpoint", "status"},
	)
)

func init() {
	prometheus.MustRegister(totalRequests)
	prometheus.MustRegister(requestDuration)
	prometheus.MustRegister(counter)
}

func main() {
	http.HandleFunc("/record-metrics", func(w http.ResponseWriter, r *http.Request) {
		endpoint := r.URL.Query().Get("endpoint")
		status := r.URL.Query().Get("status")
		counter.WithLabelValues(endpoint, status).Inc()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Metrics recorded successfully"))
	})

	http.HandleFunc("/fast", fastHandler)
	http.HandleFunc("/slow", slowHandler)

	http.Handle("/metrics", promhttp.Handler())

	fmt.Println("Server listening on :8080")
	http.ListenAndServe(":8080", nil)
}
