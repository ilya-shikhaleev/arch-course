package main

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"

	"github.com/ilya-shikhaleev/arch-course/pkg/app"
	"github.com/ilya-shikhaleev/arch-course/pkg/infrastructure/postgres"
	"github.com/ilya-shikhaleev/arch-course/pkg/infrastructure/transport"
)

func main() {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.Print("Starting the service...")

	port := os.Getenv("PORT")
	if port == "" {
		logger.Fatal("Port is not set.")
	}

	host := os.Getenv("POSTGRES_HOST")
	postgresPort := os.Getenv("POSTGRES_PORT")
	dbname := os.Getenv("POSTGRES_DB")
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	if host == "" || postgresPort == "" || dbname == "" || user == "" || password == "" {
		logger.Fatal("Postgres env is not set.")
	}
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, postgresPort, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		panic(err)
	}

	logger.Print("The service is ready to listen and serve.")
	killSignalChan := getKillSignalChan()
	srv := startServer(":"+port, logger, db)

	waitForKillSignal(killSignalChan, logger)
	_ = srv.Shutdown(context.Background())
}

func startServer(serverUrl string, logger *logrus.Logger, db *sql.DB) *http.Server {
	router := logMiddleware(httpHandler(logger, db), logger)
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

func httpHandler(logger *logrus.Logger, db *sql.DB) http.Handler {
	serverErrorLogger := &serverErrorLogger{logger}

	router := mux.NewRouter()
	router.HandleFunc("/health", healthHandler).Methods(http.MethodGet)
	router.HandleFunc("/ready", readyHandler()).Methods(http.MethodGet)
	router.HandleFunc("/info", infoHandler).Methods(http.MethodGet)

	userService := app.NewUserService(postgres.NewUserRepository(db))

	serveMux := http.NewServeMux()
	serveMux.Handle("/api/v1/", transport.MakeHandler(userService, serverErrorLogger))
	serveMux.Handle("/", router)

	return serveMux
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, "{\"status\": \"OK\"}")
}

func readyHandler() http.HandlerFunc {
	isReady := &atomic.Value{}
	isReady.Store(false)

	go func() {
		time.Sleep(5 * time.Second) // Some delay for load cache for example
		isReady.Store(true)
	}()

	return func(w http.ResponseWriter, _ *http.Request) {
		if isReady.Load().(bool) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, "{\"status\": \"READY\"}")
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
	}
}

func infoHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	hostname := os.Getenv("HOSTNAME")
	_, _ = io.WriteString(w, "{\"hostname\": \""+hostname+"\"}")
}

func logMiddleware(h http.Handler, logger *logrus.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.WithFields(logrus.Fields{
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

func waitForKillSignal(killSignalChan <-chan os.Signal, logger *logrus.Logger) {
	killSignal := <-killSignalChan
	switch killSignal {
	case os.Interrupt:
		logger.Info("got SIGINT...")
	case syscall.SIGTERM:
		logger.Info("got SIGTERM...")
	}
}

type serverErrorLogger struct {
	*logrus.Logger
}

func (l *serverErrorLogger) Log(args ...interface{}) error {
	l.Error(args...)
	return nil
}
