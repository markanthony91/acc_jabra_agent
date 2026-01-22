package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aiknow/acc_jabra_agent/internal/db"
	"github.com/aiknow/acc_jabra_agent/internal/jabra"
)

func TestEndpoints(t *testing.T) {
	// Setup
	store, _ := db.NewStore()
	monitor := jabra.NewMonitor("TEST-SERIAL", store)
	server := NewServer(monitor, store)

	t.Run("GET /api/health", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/health", nil)
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(server.handleHealth)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("status incorreto: got %v want %v", status, http.StatusOK)
		}
	})

	t.Run("GET /api/telemetry", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/telemetry", nil)
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(server.handleTelemetry)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("status incorreto: got %v want %v", status, http.StatusOK)
		}

		var resp map[string]interface{}
		json.NewDecoder(rr.Body).Decode(&resp)
		if resp["hostname"] == "" {
			t.Error("hostname não deve estar vazio")
		}
	})
}
