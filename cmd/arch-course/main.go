package main

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/ilya-shikhaleev/arch-course/pkg/app"
	"github.com/ilya-shikhaleev/arch-course/pkg/infrastructure/postgres"
	"github.com/ilya-shikhaleev/arch-course/pkg/infrastructure/transport"
)

var db *sql.DB
var readyDBCh chan *sql.DB

func main() {
	readyDBCh = make(chan *sql.DB)

	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.Print("Starting the service...")

	port := os.Getenv("PORT")
	if port == "" {
		logger.Fatal("Port is not set.")
	}

	go func() {
		db = initDB(logger)
	}()

	defer func() {
		if db != nil {
			_ = db.Close()
		}
	}()

	logger.Print("The service is ready to listen and serve.")
	killSignalChan := getKillSignalChan()
	srv := startServer(":"+port, logger)

	waitForKillSignal(killSignalChan, logger)
	_ = srv.Shutdown(context.Background())
}

func initDB(logger *logrus.Logger) *sql.DB {
	host := os.Getenv("POSTGRES_HOST")
	postgresPort := os.Getenv("POSTGRES_PORT")
	dbname := os.Getenv("POSTGRES_DB")
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	if host == "" || postgresPort == "" || dbname == "" || user == "" || password == "" {
		logger.Fatal("Postgres env is not set.")
	}

	for {
		postgresSource := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			host, postgresPort, user, password, dbname)
		db, err := sql.Open("postgres", postgresSource)
		if err != nil {
			logger.Info(errors.Wrap(err, "can't open connection to "+postgresSource))
			time.Sleep(time.Second)
			continue
		}

		err = db.Ping()
		if err != nil {
			logger.Info(errors.Wrap(err, "can't ping to "+postgresSource))
			time.Sleep(time.Second)
			continue
		}
		readyDBCh <- db
		return db
	}
}

func startServer(serverUrl string, logger *logrus.Logger) *http.Server {
	m := serveMux(logger)
	router := logMiddleware(m, logger)
	srv := &http.Server{
		Handler:      router,
		Addr:         serverUrl,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}
	go func() {
		logger.Fatal(srv.ListenAndServe())
	}()

	go func() {
		db := <-readyDBCh
		serverErrorLogger := &serverErrorLogger{logger}
		userService := app.NewUserService(postgres.NewUserRepository(db))
		m.Handle("/api/v1/", transport.MakeHandler(userService, serverErrorLogger))
	}()

	return srv
}

func serveMux(logger *logrus.Logger) *http.ServeMux {

	router := mux.NewRouter()
	router.HandleFunc("/health", healthHandler).Methods(http.MethodGet)
	router.HandleFunc("/ready", readyHandler()).Methods(http.MethodGet)
	router.HandleFunc("/info", infoHandler).Methods(http.MethodGet)

	serveMux := http.NewServeMux()
	serveMux.Handle("/", router)

	return serveMux
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, "{\"status\": \"OK\"}")
}

func readyHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		if db != nil {
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
