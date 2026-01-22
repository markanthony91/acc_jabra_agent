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

	"path/filepath"
	"runtime"
	"time"

	"github.com/aiknow/acc_jabra_agent/internal/api"
	"github.com/aiknow/acc_jabra_agent/internal/db"
	"github.com/aiknow/acc_jabra_agent/internal/jabra"
	"github.com/getlantern/systray"
	"github.com/webview/webview_go"
)

var (
	appPort string
	monitor *jabra.Monitor
	winChan = make(chan string, 1)
)

func main() {
	appPort = os.Getenv("PORT")
	if appPort == "" {
		appPort = "18888"
	}

	// 1. Inicializa Persistência (SQLite)
	store, err := db.NewStore()
	if err != nil {
		log.Fatalf("[ACC-Jabra] Erro ao iniciar banco: %v", err)
	}

	// 2. Inicializa o monitor de hardware
	monitor = jabra.NewMonitor("ABC123456789", store)

	// 3. Inicia o Servidor API/Web em background
	server := api.NewServer(monitor, store)
	go func() {
		if err := server.Start(appPort); err != nil {
			log.Fatalf("[ACC-Jabra] Erro no servidor: %v", err)
		}
	}()

	// 4. Configura Autostart (Linux)
	ensureAutostart(store)

	// 5. Worker para abrir janelas
	go func() {
		for view := range winChan {
			openWindow(view)
		}
	}()

	// 6. Inicia Systray
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetTitle("ACC Jabra")
	systray.SetTooltip("ACC Jabra Telemetry Agent")
	
	systray.SetTemplateIcon(iconData, iconData)

	mOpen := systray.AddMenuItem("Abrir ACC Jabra", "Abre a janela de telemetria")
	mDash := systray.AddMenuItem("Dashboard Completo", "Abre o dashboard no navegador")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Sair", "Encerra o agente")

	// Abre a janela inicial
	winChan <- "mini"

	go func() {
		for {
			select {
			case <-mOpen.ClickedCh:
				winChan <- "mini"
			case <-mDash.ClickedCh:
				winChan <- "full"
			case <-mQuit.ClickedCh:
				systray.Quit()
			}
		}
	}()
}

func onExit() {
	// Limpeza se necessário
}

func openWindow(view string) {
	// No Linux, webview precisa rodar na thread principal ou em uma thread de UI
	// Como systray.Run já tomou a main thread, usamos runtime.LockOSThread se necessário
	// Mas o webview_go costuma lidar bem com goroutines no Linux/GTK se inicializado corretamente
	w := webview.New(false)
	defer w.Destroy()
	
	title := "ACC Jabra"
	width, height := 340, 380
	if view == "full" {
		title = "ACC Jabra Dashboard"
		width, height = 800, 600
	}

	w.SetTitle(title)
	w.SetSize(width, height, webview.HintFixed)
	w.Navigate("http://localhost:" + appPort + "?view=" + view)
	w.Run()
}

func ensureAutostart(store *db.Store) {
	if runtime.GOOS != "linux" {
		return
	}

	home, _ := os.UserHomeDir()
	autostartDir := filepath.Join(home, ".config", "autostart")
	filepath := filepath.Join(autostartDir, "acc-jabra-agent.desktop")

	enabled := store.GetSetting("autostart", "true")
	if enabled == "false" {
		os.Remove(filepath)
		log.Printf("[ACC-Jabra] Autostart desativado e arquivo removido.")
		return
	}

	os.MkdirAll(autostartDir, 0755)
	execPath, _ := os.Executable()
	
	content := fmt.Sprintf(`[Desktop Entry]
Type=Application
Name=ACC Jabra Agent
Comment=Jabra Telemetry Agent
Exec=%s
Terminal=false
Icon=utilities-terminal
Categories=System;
`, execPath)

	filepath := filepath.Join(autostartDir, "acc-jabra-agent.desktop")
	os.WriteFile(filepath, []byte(content), 0644)
	log.Printf("[ACC-Jabra] Autostart configurado em: %s", filepath)
}

// Icone minimalista em base64 (um círculo azul simples)
var iconData = []byte{
	0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
	0x00, 0x00, 0x00, 0x10, 0x00, 0x00, 0x00, 0x10, 0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x91, 0x68,
	0x36, 0x00, 0x00, 0x00, 0x19, 0x49, 0x44, 0x41, 0x54, 0x78, 0xda, 0x63, 0x64, 0xf8, 0xff, 0xbf,
	0x1e, 0x03, 0x06, 0x18, 0xec, 0x19, 0x10, 0x12, 0x03, 0x20, 0x54, 0x00, 0x00, 0x2e, 0x0e, 0x01,
	0x01, 0x8b, 0x21, 0x4d, 0x25, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae, 0x42, 0x60,
	0x82,
}
