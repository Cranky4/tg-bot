package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	labelNames []string = []string{"command"}

	TotalRequestCounter                       *prometheus.CounterVec
	GRPCReqeustTotalCounter                   *prometheus.CounterVec
	MessageBrokerMessagesProducesTotalCounter *prometheus.CounterVec
	ResponseTimeSummary                       *prometheus.SummaryVec
	MessageBrokerMessagesConsumedTotalCounter *prometheus.CounterVec
)

func init() {
	TotalRequestCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "tg_bot",
			Subsystem: "tg_client",
			Help:      "Total count of requests",
			Name:      "requests_total",
		},
		labelNames,
	)

	GRPCReqeustTotalCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "tg_bot",
			Subsystem: "tg_client",
			Help:      "Total count of GRPC requests",
			Name:      "grpc_requests_total",
		},
		labelNames,
	)

	MessageBrokerMessagesProducesTotalCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "tg_bot",
			Subsystem: "tg_client",
			Help:      "Total count of Message Broker's messages produced",
			Name:      "message_broker_messages_produced_total",
		},
		[]string{"queue"},
	)

	ResponseTimeSummary = promauto.NewSummaryVec(
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

	MessageBrokerMessagesConsumedTotalCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "tg_bot",
			Subsystem: "reporter",
			Help:      "Total count of Message Broker's messages consumed",
			Name:      "message_broker_messages_consumed_total",
		},
		[]string{"queue"},
	)
}
