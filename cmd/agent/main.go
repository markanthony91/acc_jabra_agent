package main

import (
	"log"
	"os"
	"time"

	"github.com/aiknow/acc_jabra_agent/internal/api"
	"github.com/aiknow/acc_jabra_agent/internal/db"
	"github.com/aiknow/acc_jabra_agent/internal/jabra"
	"github.com/webview/webview_go"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "18888"
	}

	// 1. Inicializa Persistência (SQLite)
	store, err := db.NewStore()
	if err != nil {
		log.Fatalf("[ACC-Jabra] Erro ao iniciar banco: %v", err)
	}

	// 2. Inicializa o monitor de hardware com acesso ao banco
	monitor := jabra.NewMonitor("ABC123456789", store)

	// 3. Inicia o Servidor API/Web em background
	server := api.NewServer(monitor, store)
	go func() {
		if err := server.Start(port); err != nil {
			log.Fatalf("[ACC-Jabra] Erro no servidor: %v", err)
		}
	}()

	time.Sleep(500 * time.Millisecond)

	// 4. Lança a Janela Nativa
	w := webview.New(false)
	defer w.Destroy()
	w.SetTitle("ACC Jabra Telemetry")
	w.SetSize(380, 520, webview.HintFixed)
	w.Navigate("http://localhost:" + port)
	w.Run()
}
