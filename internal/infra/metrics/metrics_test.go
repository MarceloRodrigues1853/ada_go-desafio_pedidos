package metrics_test

import (
	"testing"

	"pedidos/internal/infra/metrics"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func gatherFamily(t *testing.T, name string) *dto.MetricFamily {
	t.Helper()
	families, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("erro ao coletar métricas: %v", err)
	}
	for _, mf := range families {
		if mf.GetName() == name {
			return mf
		}
	}
	t.Fatalf("métrica %s não registrada", name)
	return nil
}

func counterTotal(t *testing.T, name string) float64 {
	t.Helper()
	families, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("erro ao coletar métricas: %v", err)
	}
	for _, mf := range families {
		if mf.GetName() == name {
			var total float64
			for _, m := range mf.GetMetric() {
				total += m.GetCounter().GetValue()
			}
			return total
		}
	}
	return 0
}

func sampleCount(t *testing.T, name string) uint64 {
	t.Helper()
	families, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("erro ao coletar métricas: %v", err)
	}
	for _, mf := range families {
		if mf.GetName() == name {
			var total uint64
			for _, m := range mf.GetMetric() {
				total += m.GetHistogram().GetSampleCount()
			}
			return total
		}
	}
	return 0
}

func labelValue(m *dto.Metric, name string) string {
	for _, lp := range m.GetLabel() {
		if lp.GetName() == name {
			return lp.GetValue()
		}
	}
	return ""
}

func TestInitMetricsRegistersCollectors(t *testing.T) {
	metrics.InitMetrics()
	metrics.InitMetrics() // deve ser idempotente (sem pânico)

	metrics.IncOrdersCreated()
	metrics.IncPaymentsProcessed("PAID")
	metrics.IncMessagesDLQ()
	metrics.ObserveOrderProcessing(0.100)

	expected := []string{
		"orders_created_total",
		"payments_processed_total",
		"messages_dlq_total",
		"order_processing_duration_seconds",
	}
	for _, name := range expected {
		gatherFamily(t, name)
	}
}

func TestIncOrdersCreated(t *testing.T) {
	metrics.InitMetrics()
	before := counterTotal(t, "orders_created_total")
	metrics.IncOrdersCreated()
	if got := counterTotal(t, "orders_created_total"); got != before+1 {
		t.Errorf("orders_created_total: esperado %v, obtido %v", before+1, got)
	}
}

func TestIncPaymentsProcessed(t *testing.T) {
	metrics.InitMetrics()
	before := counterTotal(t, "payments_processed_total")

	metrics.IncPaymentsProcessed("PAID")
	metrics.IncPaymentsProcessed("FAILED")

	if got := counterTotal(t, "payments_processed_total"); got != before+2 {
		t.Errorf("payments_processed_total: esperado %v, obtido %v", before+2, got)
	}

	mf := gatherFamily(t, "payments_processed_total")
	statuses := map[string]bool{}
	for _, m := range mf.GetMetric() {
		statuses[labelValue(m, "status")] = true
		if m.GetCounter().GetValue() < 1 {
			t.Errorf("status %s sem incremento", labelValue(m, "status"))
		}
	}
	if !statuses["PAID"] || !statuses["FAILED"] {
		t.Errorf("status esperados PAID e FAILED, obtidos %v", statuses)
	}
}

func TestIncMessagesDLQ(t *testing.T) {
	metrics.InitMetrics()
	before := counterTotal(t, "messages_dlq_total")
	metrics.IncMessagesDLQ()
	if got := counterTotal(t, "messages_dlq_total"); got != before+1 {
		t.Errorf("messages_dlq_total: esperado %v, obtido %v", before+1, got)
	}
}

func TestObserveOrderProcessing(t *testing.T) {
	metrics.InitMetrics()
	before := sampleCount(t, "order_processing_duration_seconds")

	metrics.ObserveOrderProcessing(0.100)
	metrics.ObserveOrderProcessing(0.250)

	if got := sampleCount(t, "order_processing_duration_seconds"); got != before+2 {
		t.Errorf("order_processing_duration_seconds: esperado %v amostras, obtido %v", before+2, got)
	}
}
