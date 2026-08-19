package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	OrdersCreatedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "orders_created_total",
			Help: "Total de pedidos criados com sucesso.",
		},
	)

	PaymentsProcessedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "payments_processed_total",
			Help: "Total de pagamentos processados, classificados por status.",
		},
		[]string{"status"}, // "PAID" | "FAILED"
	)

	MessagesDLQTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "messages_dlq_total",
			Help: "Total de mensagens rejeitadas e encaminhadas para a Dead Letter Queue (DLQ).",
		},
	)

	OrderProcessingDurationSeconds = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "order_processing_duration_seconds",
			Help:    "Histograma de latência do processamento de pedidos em segundos.",
			Buckets: prometheus.DefBuckets,
		},
	)
)

// InitMetrics registra as métricas no prometheus.DefaultRegisterer.
// É idempotente: registros duplicados são ignorados.
func InitMetrics() {
	for _, collector := range []prometheus.Collector{
		OrdersCreatedTotal,
		PaymentsProcessedTotal,
		MessagesDLQTotal,
		OrderProcessingDurationSeconds,
	} {
		if err := prometheus.Register(collector); err != nil {
			if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
				panic(err)
			}
		}
	}
}

func IncOrdersCreated() {
	OrdersCreatedTotal.Inc()
}

func IncPaymentsProcessed(status string) {
	PaymentsProcessedTotal.WithLabelValues(status).Inc()
}

func IncMessagesDLQ() {
	MessagesDLQTotal.Inc()
}

func ObserveOrderProcessing(durationSeconds float64) {
	OrderProcessingDurationSeconds.Observe(durationSeconds)
}
