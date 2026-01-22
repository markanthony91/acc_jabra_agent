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
	http.HandleFunc("/api/logs", s.handleLogs)
	http.HandleFunc("/api/config", s.handleConfig)
	http.HandleFunc("/api/health", s.handleHealth)

	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/", fs)

	return http.ListenAndServe(":"+port, nil)
}

func (s *Server) handleLogs(w http.ResponseWriter, r *http.Request) {
	logs, err := s.store.GetLogs(100)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var cfg map[string]string
		if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		for k, v := range cfg {
			s.store.SaveSetting(k, v)
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	// GET: Retorna configs atuais
	config := map[string]string{
		"operator_name": s.store.GetSetting("operator_name", "Operador 01"),
		"custom_color":  s.store.GetSetting("custom_color", "#2196F3"),
		"autostart":     s.store.GetSetting("autostart", "true"),
		"show_tray":      s.store.GetSetting("show_tray", "true"),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
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
