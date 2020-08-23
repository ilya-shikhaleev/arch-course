package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
	"github.com/ispringteam/go-patterns/infrastructure/jsonlog"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"

	"github.com/ilya-shikhaleev/arch-course/pkg/common/amqp"
	"github.com/ilya-shikhaleev/arch-course/pkg/popular/infrastructure/handler"
	"github.com/ilya-shikhaleev/arch-course/pkg/popular/infrastructure/postgres"
	"github.com/ilya-shikhaleev/arch-course/pkg/popular/infrastructure/transport"
)

var db *sql.DB
var readyDBCh chan *sql.DB

var redisClient *redis.Client
var readyRedisCh chan *redis.Client

var amqpConnection *amqp.Connection
var readyOrderDomainEventChannelCh chan amqp.Channel

func main() {
	readyDBCh = make(chan *sql.DB)
	readyRedisCh = make(chan *redis.Client)
	readyOrderDomainEventChannelCh = make(chan amqp.Channel)
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.Print("Starting the service...")

	port := os.Getenv("PORT")
	if port == "" {
		logger.Fatal("Port is not set.")
	}

	go func() {
		amqpConnection = initRabbitMQ(logger)
	}()
	defer func() {
		if amqpConnection != nil {
			_ = amqpConnection.Close()
		}
	}()

	go func() {
		redisClient = initRedis(logger)
	}()
	defer func() {
		if redisClient != nil {
			_ = redisClient.Close()
		}
	}()

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
		db := <-readyDBCh
		redisClient := <-readyRedisCh
		serverErrorLogger := &serverErrorLogger{logger}
		repo := postgres.NewPopularRepository(db, redisClient)

		go func() { // TODO: move it out of there
			orderDomainEventChannel := <-readyOrderDomainEventChannelCh
			logger.Info("reading events is prepared")
			for data := range orderDomainEventChannel.Receive() {
				logger.Info("new event", data)
				var req handler.OnBuyProductsRequest
				if err := json.Unmarshal([]byte(data), &req); err != nil {
					logger.Info(err, "invalid order paid event")
					continue
				}
				logger.Info("event data", req)
				if err := handler.OnBuyProducts(req, repo); err != nil {
					logger.Info(err, "can't process order paid event")
				}
			}
		}()

		m.Handle("/api/v1/", transport.MakeHandler(repo, serverErrorLogger))
	}()

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
	router.HandleFunc("/ready", readyHandler()).Methods(http.MethodGet)
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

func initDB(logger *logrus.Logger) *sql.DB {
	host := os.Getenv("POSTGRES_HOST")
	postgresPort := os.Getenv("POSTGRES_PORT")
	dbname := os.Getenv("POSTGRES_DB")
	dbUser := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	if host == "" || postgresPort == "" || dbname == "" || dbUser == "" || password == "" {
		logger.Fatal("Postgres env is not set.")
	}

	for {
		postgresSource := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			host, postgresPort, dbUser, password, dbname)
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

func initRedis(logger *logrus.Logger) *redis.Client {
	host := os.Getenv("REDIS_HOST")
	port := os.Getenv("REDIS_PORT")
	password := os.Getenv("REDIS_PASSWORD")
	if host == "" || port == "" || password == "" {
		logger.Fatal("Redis env is not set.")
	}

	for {
		source := fmt.Sprintf("%s:%v", host, port)
		redisClient := redis.NewClient(&redis.Options{
			Addr:     source,
			Password: password,
		})
		if redisClient == nil {
			logger.Info("can't open connection to " + source)
			time.Sleep(time.Second)
			continue
		}
		_, err := redisClient.Ping().Result()
		if err != nil {
			logger.Info(errors.Wrap(err, "can't ping to "+source))
			time.Sleep(time.Second)
			continue
		}
		readyRedisCh <- redisClient
		return redisClient
	}
}

func initRabbitMQ(logger *logrus.Logger) *amqp.Connection {
	host := os.Getenv("RABBITMQ_HOST")
	user := os.Getenv("RABBITMQ_USER")
	password := os.Getenv("RABBITMQ_PASSWORD")
	if host == "" || user == "" || password == "" {
		logger.Fatal("rabbitmq env is not set.")
	}
	l := jsonlog.NewLogger(&jsonlog.Config{
		Level:   jsonlog.InfoLevel,
		AppName: "order",
	})

	for {
		amqpConnection := amqp.NewAMQPConnection(&amqp.Config{Host: host, User: user, Password: password}, l)
		ch := amqp.NewOrderDomainEventsChannel()
		amqpConnection.AddChannel(ch)
		err := amqpConnection.Start()
		if err != nil {
			logger.Info(errors.Wrap(err, "can't open connection to amqp"))
			time.Sleep(time.Second)
			continue
		}

		readyOrderDomainEventChannelCh <- ch
		return amqpConnection
	}
}

type serverErrorLogger struct {
	*logrus.Logger
}

func (l *serverErrorLogger) Log(args ...interface{}) error {
	l.Error(args...)
	return nil
}
