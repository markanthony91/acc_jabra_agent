package api

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/aiknow/acc_jabra_agent/internal/db"
	"github.com/aiknow/acc_jabra_agent/internal/jabra"
)

type Server struct {
	monitor *jabra.Monitor
	store   *db.Store
}

func NewServer(m *jabra.Monitor, s *db.Store) *Server {
	return &Server{monitor: m, store: s}
}

func (s *Server) Start(port string) error {
	http.HandleFunc("/api/telemetry", s.handleTelemetry)
	http.HandleFunc("/api/history/battery", s.handleBatteryHistory)
	http.HandleFunc("/api/health", s.handleHealth)

	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/", fs)

	return http.ListenAndServe(":"+port, nil)
}

func (s *Server) handleBatteryHistory(w http.ResponseWriter, r *http.Request) {
	history, err := s.store.GetBatteryHistory(50)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

func (s *Server) handleTelemetry(w http.ResponseWriter, r *http.Request) {
	data := s.monitor.GetTelemetry()
	w.Header().Set("Content-Type", "application/json")
	hostname, _ := os.Hostname()
	json.NewEncoder(w).Encode(map[string]interface{}{"hostname": hostname, "data": data})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
