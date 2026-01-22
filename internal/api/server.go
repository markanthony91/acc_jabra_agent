package api

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"github.com/aiknow/acc_jabra_agent/internal/jabra"
)

type Server struct {
	monitor *jabra.Monitor
}

func NewServer(m *jabra.Monitor) *Server {
	return &Server{monitor: m}
}

func (s *Server) Start(port string) error {
	// API Handlers
	http.HandleFunc("/api/telemetry", s.handleTelemetry)
	http.HandleFunc("/api/health", s.handleHealth)

	// Static Files Handler (Dashboard)
	// Servindo a pasta /public na raiz /
	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/", fs)

	return http.ListenAndServe(":"+port, nil)
}

func (s *Server) handleTelemetry(w http.ResponseWriter, r *http.Request) {
	data := s.monitor.GetTelemetry()
	
	w.Header().Set("Content-Type", "application/json")
	hostname, _ := os.Hostname()
	
	response := struct {
		Hostname string      `json:"hostname"`
		Data     interface{} `json:"data"`
	}{
		Hostname: hostname,
		Data:     data,
	}

	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}