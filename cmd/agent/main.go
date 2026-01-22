package main

import (
	"log"
	"os"
	"time"

	"github.com/aiknow/acc_jabra_agent/internal/api"
	"github.com/aiknow/acc_jabra_agent/internal/jabra"
	"github.com/webview/webview_go"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "18888"
	}

	// 1. Inicializa o monitor de hardware
	monitor := jabra.NewMonitor("ABC123456789")

	// 2. Inicia o Servidor API/Web em uma Goroutine (background)
	server := api.NewServer(monitor)
	go func() {
		log.Printf("[ACC-Jabra] Servidor rodando em http://localhost:%s", port)
		if err := server.Start(port); err != nil {
			log.Fatalf("[ACC-Jabra] Erro no servidor: %v", err)
		}
	}()

	// Aguarda um momento para o servidor subir
	time.Sleep(500 * time.Millisecond)

	// 3. Lança a Janela Nativa (Desktop App)
	w := webview.New(false)
	defer w.Destroy()
	w.SetTitle("ACC Jabra Telemetry")
	w.SetSize(380, 520, webview.HintFixed) // Tamanho fixo para parecer um widget
	w.Navigate("http://localhost:" + port)

	log.Printf("[ACC-Jabra] Janela desktop iniciada.")
	w.Run()
}