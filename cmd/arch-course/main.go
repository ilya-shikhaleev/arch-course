package main

import (
	"context"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	log.Print("Starting the service...")

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("Port is not set.")
	}

	log.Print("The service is ready to listen and serve.")
	killSignalChan := getKillSignalChan()
	srv := startServer(":" + port)

	waitForKillSignal(killSignalChan)
	srv.Shutdown(context.Background())
}

func startServer(serverUrl string) *http.Server {
	router := logMiddleware(router())
	srv := &http.Server{Addr: serverUrl, Handler: router}
	go func() {
		log.Fatal(srv.ListenAndServe())
	}()

	return srv
}

func router() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/health", healthHandler).Methods(http.MethodGet)
	r.HandleFunc("/info", infoHandler).Methods(http.MethodGet)
	return r
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "{\"status\": \"OK\"}")
}

func infoHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	hostname := os.Getenv("HOSTNAME")
	io.WriteString(w, "{\"hostname\": \""+hostname+"\"}")
}

func logMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.WithFields(log.Fields{
			"method":     r.Method,
			"url":        r.URL,
			"remoteAddr": r.RemoteAddr,
			"userAgent":  r.UserAgent(),
		}).Info("got a new request")
		h.ServeHTTP(w, r)
	})
}

func getKillSignalChan() chan os.Signal {
	osKillSignalChan := make(chan os.Signal, 1)
	signal.Notify(osKillSignalChan, os.Interrupt, syscall.SIGTERM)
	return osKillSignalChan
}

func waitForKillSignal(killSignalChan <-chan os.Signal) {
	killSignal := <-killSignalChan
	switch killSignal {
	case os.Interrupt:
		log.Info("got SIGINT...")
	case syscall.SIGTERM:
		log.Info("got SIGTERM...")
	}
}
