package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPaymentsHTTPPort(t *testing.T) {
	t.Run("PORT tem precedencia em cloud", func(t *testing.T) {
		t.Setenv("PORT", "10000")
		t.Setenv("METRICS_PORT", "9091")
		if got := paymentsHTTPPort(); got != "10000" {
			t.Fatalf("paymentsHTTPPort() = %q, esperado %q", got, "10000")
		}
	})

	t.Run("usa METRICS_PORT localmente", func(t *testing.T) {
		t.Setenv("PORT", "")
		t.Setenv("METRICS_PORT", "9191")
		if got := paymentsHTTPPort(); got != "9191" {
			t.Fatalf("paymentsHTTPPort() = %q, esperado %q", got, "9191")
		}
	})

	t.Run("usa porta padrao", func(t *testing.T) {
		t.Setenv("PORT", "")
		t.Setenv("METRICS_PORT", "")
		if got := paymentsHTTPPort(); got != defaultMetricsPort {
			t.Fatalf("paymentsHTTPPort() = %q, esperado %q", got, defaultMetricsPort)
		}
	})
}

func TestPaymentsHealth(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/health", nil)

	paymentsHTTPHandler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, esperado %d", recorder.Code, http.StatusOK)
	}
	if contentType := recorder.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("Content-Type = %q, esperado application/json", contentType)
	}
	if body := strings.TrimSpace(recorder.Body.String()); body != `{"status":"ok"}` {
		t.Fatalf("body = %q, esperado JSON de saúde", body)
	}
}
