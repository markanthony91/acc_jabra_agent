package main

import (
	"log"
	"os"

	"github.com/aiknow/acc_jabra_agent/internal/api"
	"github.com/aiknow/acc_jabra_agent/internal/jabra"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "18888"
	}

	log.Printf("[ACC-Jabra] Inicializando Agente de Telemetria...")

	// Inicializa o monitor com um serial fictício ou do hardware
	monitor := jabra.NewMonitor("ABC123456789")

	// Inicializa o servidor API
	server := api.NewServer(monitor)

	log.Printf("[ACC-Jabra] Servidor rodando na porta %s", port)
	if err := server.Start(port); err != nil {
		log.Fatalf("[ACC-Jabra] Erro ao iniciar servidor: %v", err)
	}
}
