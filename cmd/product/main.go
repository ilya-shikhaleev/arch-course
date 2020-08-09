package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.Print("Starting the service...")

	port := os.Getenv("PORT")
	if port == "" {
		logger.Fatal("Port is not set.")
	}

	logger.Print("The service is ready to listen and serve.")
	killSignalChan := getKillSignalChan()
	srv := startServer(":"+port, logger)

	waitForKillSignal(killSignalChan, logger)
	_ = srv.Shutdown(context.Background())
}

func startServer(serverUrl string, logger *logrus.Logger) *http.Server {
	histogram := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "app_request_latency_seconds",
		Help:    "Application Request Latency.",
		Buckets: prometheus.DefBuckets,
	}, []string{"endpoint", "method", "status"})
	// Registering the defined metric with Prometheus
	_ = prometheus.Register(histogram)
	counter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "app_request_count",
		Help: "Application Request Count.",
	}, []string{"endpoint", "method", "status"})
	// Registering the defined metric with Prometheus
	_ = prometheus.Register(counter)

	m := serveMux()
	router := metricsMiddleware(logMiddleware(m, logger), histogram, counter)
	srv := &http.Server{
		Handler:      router,
		Addr:         serverUrl,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}
	go func() {
		logger.Fatal(srv.ListenAndServe())
	}()

	return srv
}

func logMiddleware(h http.Handler, logger *logrus.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		statusWriter := statusWriter{ResponseWriter: w}
		h.ServeHTTP(&statusWriter, r)
		if r.URL.Path != "/health" && r.URL.Path != "/ready" {
			logger.WithFields(logrus.Fields{
				"method":     r.Method,
				"url":        r.URL,
				"remoteAddr": r.RemoteAddr,
				"userAgent":  r.UserAgent(),
				"code":       statusWriter.status,
			}).Info("got a new request")
		}
	})
}

func serveMux() *http.ServeMux {
	router := mux.NewRouter()
	router.HandleFunc("/health", healthHandler).Methods(http.MethodGet)
	router.HandleFunc("/ready", readyHandler).Methods(http.MethodGet)
	router.HandleFunc("/info", infoHandler).Methods(http.MethodGet)
	router.Handle("/metrics", promhttp.Handler())

	serveMux := http.NewServeMux()
	serveMux.Handle("/", router)

	return serveMux
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, "{\"status\": \"OK\"}")
}

func readyHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, "{\"status\": \"READY\"}")
}

func infoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-User-Id") == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	response := struct {
		ID        string `json:"id"`
		Login     string `json:"login"`
		Email     string `json:"email"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}{
		r.Header.Get("X-User-Id"),
		r.Header.Get("X-Login"),
		r.Header.Get("X-Email"),
		r.Header.Get("X-First-Name"),
		r.Header.Get("X-Last-Name"),
	}

	result, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, string(result))
}

func metricsMiddleware(h http.Handler, histogram *prometheus.HistogramVec, counter *prometheus.CounterVec) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		code := http.StatusBadRequest

		defer func() {
			httpDuration := time.Since(start)
			histogram.WithLabelValues(r.RequestURI, r.Method, fmt.Sprintf("%d", code)).Observe(httpDuration.Seconds())
			counter.WithLabelValues(r.RequestURI, r.Method, fmt.Sprintf("%d", code)).Inc()
		}()

		statusWriter := statusWriter{ResponseWriter: w}
		h.ServeHTTP(&statusWriter, r)
		code = statusWriter.status
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.ResponseWriter.Write(b)
}

func getKillSignalChan() chan os.Signal {
	osKillSignalChan := make(chan os.Signal, 1)
	signal.Notify(osKillSignalChan, os.Interrupt, syscall.SIGTERM)
	return osKillSignalChan
}

func waitForKillSignal(killSignalChan <-chan os.Signal, logger *logrus.Logger) {
	killSignal := <-killSignalChan
	switch killSignal {
	case os.Interrupt:
		logger.Info("got SIGINT...")
	case syscall.SIGTERM:
		logger.Info("got SIGTERM...")
	}
}
