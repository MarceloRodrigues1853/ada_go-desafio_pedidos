package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pedidos/internal/infra/metrics"

	"github.com/go-chi/chi/v5"
)

// TestMetricsEndpoint verifica que GET /metrics responde 200, com Content-Type
// compatível com Prometheus e contendo as métricas registradas pela aplicação.
func TestMetricsEndpoint(t *testing.T) {
	metrics.InitMetrics()                // mesmo registro que main() faz em produção
	metrics.IncPaymentsProcessed("PAID") // garante a emissão do CounterVec no /metrics

	router := newRouter(nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("GET /metrics: esperado HTTP 200, obtido %d", rr.Code)
	}

	contentType := rr.Header().Get("Content-Type")
	if !strings.HasPrefix(contentType, "text/plain") {
		t.Errorf("GET /metrics: Content-Type incompatível com Prometheus: %q", contentType)
	}

	body := rr.Body.String()
	for _, name := range []string{
		"orders_created_total",
		"payments_processed_total",
		"messages_dlq_total",
		"order_processing_duration_seconds",
	} {
		if !strings.Contains(body, name) {
			t.Errorf("GET /metrics: métrica %s não presente na resposta", name)
		}
	}
}

// TestMetricsEndpointDoesNotInterfere verifica que o endpoint /metrics não
// interfere no roteamento: rotas desconhecidas continuam retornando 404.
func TestMetricsEndpointDoesNotInterfere(t *testing.T) {
	router := newRouter(nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/rota-inexistente", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("GET /rota-inexistente: esperado HTTP 404, obtido %d", rr.Code)
	}
}

// TestExistingRoutesStillRegistered verifica que todas as rotas existentes
// continuam registradas no roteador após a inclusão de /metrics.
func TestExistingRoutesStillRegistered(t *testing.T) {
	router := newRouter(nil, nil, nil)

	cases := []struct{ method, path string }{
		{"GET", "/metrics"},
		{"POST", "/clientes"},
		{"GET", "/clientes"},
		{"GET", "/clientes/123e4567-e89b-12d3-a456-426614174000"},
		{"POST", "/produtos"},
		{"GET", "/produtos"},
		{"GET", "/produtos/sku-001"},
		{"POST", "/pedidos"},
		{"GET", "/pedidos"},
		{"GET", "/pedidos/123e4567-e89b-12d3-a456-426614174000"},
		{"POST", "/pedidos/123e4567-e89b-12d3-a456-426614174000/pagar"},
		{"POST", "/pedidos/123e4567-e89b-12d3-a456-426614174000/cancelar"},
	}

	for _, tc := range cases {
		if !router.Match(chi.NewRouteContext(), tc.method, tc.path) {
			t.Errorf("rota %s %s não está mais registrada", tc.method, tc.path)
		}
	}

	if router.Match(chi.NewRouteContext(), http.MethodGet, "/rota-inexistente") {
		t.Error("rota desconhecida deveria continuar sem match")
	}
}
