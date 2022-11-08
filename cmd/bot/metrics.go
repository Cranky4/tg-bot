package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var labelNames []string = []string{"command"}

func initTotalCounter() *prometheus.CounterVec {
	return promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "tg_bot",
			Subsystem: "tg_client",
			Help:      "Total count of requests",
			Name:      "requests_total",
		},
		labelNames,
	)
}

func initGRPCTotalCounter() *prometheus.CounterVec {
	return promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "tg_bot",
			Subsystem: "tg_client",
			Help:      "Total count of GRPC requests",
			Name:      "grpc_requests_total",
		},
		labelNames,
	)
}

func initMessageBrokerMessagesProducesTotalCounter() *prometheus.CounterVec {
	return promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "tg_bot",
			Subsystem: "tg_client",
			Help:      "Total count of Message Broker's messages produced",
			Name:      "message_broker_messages_produced_total",
		},
		labelNames,
	)
}

func initResponseTime() *prometheus.SummaryVec {
	return promauto.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "tg_bot",
			Subsystem: "tg_client",
			Help:      "Response time of commands",
			Name:      "response_time_milliseconds",
			Objectives: map[float64]float64{
				0.5:  50,
				0.9:  10,
				0.99: 1,
			},
		}, labelNames,
	)
}

func startMetricsHTTPServer(url string, port int) error {
	http.Handle(url, promhttp.Handler())

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		ReadHeaderTimeout: 3 * time.Second,
	}

	err := server.ListenAndServe()
	if err != nil {
		return errors.Wrap(err, "ошибка старта сервера метрик")
	}

	return nil
}
