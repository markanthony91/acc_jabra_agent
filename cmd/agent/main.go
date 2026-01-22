package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/aiknow/acc_jabra_agent/internal/actions"
	"github.com/aiknow/acc_jabra_agent/internal/api"
	"github.com/aiknow/acc_jabra_agent/internal/autostart"
	"github.com/aiknow/acc_jabra_agent/internal/db"
	"github.com/aiknow/acc_jabra_agent/internal/jabra"
	"github.com/aiknow/acc_jabra_agent/internal/security"
	"github.com/aiknow/acc_jabra_agent/internal/socket"
	"github.com/getlantern/systray"
	"github.com/webview/webview_go"
)

// App contém todas as dependências da aplicação
type App struct {
	Port     string
	Store    *db.Store
	Monitor  *jabra.Monitor
	Server   *api.Server
	Socket   *socket.Client
	Executor *actions.Executor
	Whitelist *security.Whitelist
	WinChan  chan string
}

// SocketConfig representa a configuração do Socket.IO
type SocketConfig struct {
	Host  string `json:"host"`
	Port  int    `json:"port"`
	Token string `json:"token"`
	Ramal string `json:"ramal"`
}

var app *App

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println("[ACC-Jabra] Iniciando aplicação...")

	app = &App{
		Port:    getEnvOrDefault("PORT", "18888"),
		WinChan: make(chan string, 1),
	}

	// 1. Inicializa Persistência (SQLite)
	var err error
	app.Store, err = db.NewStore()
	if err != nil {
		log.Fatalf("[ACC-Jabra] Erro ao iniciar banco: %v", err)
	}
	log.Println("[ACC-Jabra] Banco de dados inicializado")

	// 2. Inicializa Whitelist de dispositivos
	whitelistPath := getConfigPath("allowed_devices.json")
	app.Whitelist, err = security.NewWhitelist(whitelistPath)
	if err != nil {
		log.Printf("[ACC-Jabra] Aviso: Whitelist não carregado: %v", err)
	}

	// 3. Inicializa o monitor de hardware
	serialNumber := app.Store.GetSetting("device_serial", "")
	app.Monitor = jabra.NewMonitor(serialNumber, app.Store)
	log.Println("[ACC-Jabra] Monitor de hardware inicializado")

	// 4. Inicializa executor de ações (keymap)
	keymapPath := getConfigPath("keymap.json")
	app.Executor, err = actions.NewExecutor(keymapPath)
	if err != nil {
		log.Printf("[ACC-Jabra] Aviso: Executor usando keymap padrão: %v", err)
	}

	// 5. Inicializa cliente Socket.IO
	socketConfig := loadSocketConfig()
	if socketConfig.Host != "" {
		app.Socket = socket.NewClient(socket.Config{
			Host:  socketConfig.Host,
			Port:  socketConfig.Port,
			Token: socketConfig.Token,
			Ramal: socketConfig.Ramal,
		})

		// Conecta executor ao socket
		app.Executor.SetSocketEmitter(app.Socket)

		// Registra callbacks do Socket
		registerSocketCallbacks()

		// Conecta ao servidor
		go func() {
			if err := app.Socket.Connect(); err != nil {
				log.Printf("[ACC-Jabra] Erro ao conectar Socket.IO: %v", err)
			}
		}()
	}

	// 6. Inicia o Servidor API/Web em background
	app.Server = api.NewServer(app.Monitor, app.Store)
	go func() {
		log.Printf("[ACC-Jabra] Iniciando servidor na porta %s", app.Port)
		if err := app.Server.Start(app.Port); err != nil {
			log.Fatalf("[ACC-Jabra] Erro no servidor: %v", err)
		}
	}()

	// 7. Configura Autostart
	setupAutostart()

	// 8. Worker para abrir janelas
	go windowWorker()

	// 9. Inicia Systray (bloqueia)
	log.Println("[ACC-Jabra] Iniciando System Tray...")
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetTitle("ACC Jabra")
	systray.SetTooltip("ACC Jabra Telemetry Agent")
	systray.SetTemplateIcon(iconData, iconData)

	mOpen := systray.AddMenuItem("Abrir ACC Jabra", "Abre a janela de telemetria")
	mDash := systray.AddMenuItem("Dashboard Completo", "Abre o dashboard completo")
	systray.AddSeparator()

	mAutostart := systray.AddMenuItemCheckbox("Iniciar com o Sistema", "Ativa/desativa início automático", autostart.IsEnabled())
	systray.AddSeparator()

	mQuit := systray.AddMenuItem("Sair", "Encerra o agente")

	// Abre a janela inicial
	app.WinChan <- "mini"

	go func() {
		for {
			select {
			case <-mOpen.ClickedCh:
				app.WinChan <- "mini"

			case <-mDash.ClickedCh:
				app.WinChan <- "full"

			case <-mAutostart.ClickedCh:
				toggleAutostart(mAutostart)

			case <-mQuit.ClickedCh:
				cleanup()
				systray.Quit()
			}
		}
	}()
}

func onExit() {
	log.Println("[ACC-Jabra] Encerrando aplicação...")
	cleanup()
}

func cleanup() {
	// Fecha Socket.IO
	if app.Socket != nil {
		app.Socket.Disconnect()
	}

	// Para whitelist enforcement
	if app.Whitelist != nil {
		app.Whitelist.StopEnforcement()
	}

	// Fecha canal de janelas
	close(app.WinChan)
}

func windowWorker() {
	for view := range app.WinChan {
		openWindow(view)
	}
}

func openWindow(view string) {
	w := webview.New(false)
	if w == nil {
		log.Println("[ACC-Jabra] Erro ao criar webview")
		return
	}
	defer w.Destroy()

	title := "ACC Jabra"
	width, height := 340, 380

	if view == "full" {
		title = "ACC Jabra Dashboard"
		width, height = 800, 600
	}

	w.SetTitle(title)
	w.SetSize(width, height, webview.HintFixed)
	w.Navigate("http://localhost:" + app.Port + "?view=" + view)
	w.Run()
}

func setupAutostart() {
	enabled := app.Store.GetSetting("autostart", "true")

	if enabled == "true" {
		execPath, err := os.Executable()
		if err != nil {
			log.Printf("[ACC-Jabra] Erro ao obter caminho do executável: %v", err)
			return
		}

		if err := autostart.Enable(execPath); err != nil {
			log.Printf("[ACC-Jabra] Erro ao configurar autostart: %v", err)
		} else {
			log.Println("[ACC-Jabra] Autostart configurado")
		}
	} else {
		if err := autostart.Disable(); err != nil {
			log.Printf("[ACC-Jabra] Erro ao desabilitar autostart: %v", err)
		}
	}
}

func toggleAutostart(menuItem *systray.MenuItem) {
	if autostart.IsEnabled() {
		if err := autostart.Disable(); err != nil {
			log.Printf("[ACC-Jabra] Erro ao desabilitar autostart: %v", err)
			return
		}
		menuItem.Uncheck()
		app.Store.SetSetting("autostart", "false")
		log.Println("[ACC-Jabra] Autostart desabilitado")
	} else {
		execPath, _ := os.Executable()
		if err := autostart.Enable(execPath); err != nil {
			log.Printf("[ACC-Jabra] Erro ao habilitar autostart: %v", err)
			return
		}
		menuItem.Check()
		app.Store.SetSetting("autostart", "true")
		log.Println("[ACC-Jabra] Autostart habilitado")
	}
}

func registerSocketCallbacks() {
	if app.Socket == nil {
		return
	}

	app.Socket.OnConnectionChange(func(connected bool) {
		if connected {
			log.Println("[ACC-Jabra] Socket.IO conectado")
		} else {
			log.Println("[ACC-Jabra] Socket.IO desconectado")
		}
	})

	app.Socket.OnNotificarCarro(func(temCarro bool) {
		log.Printf("[ACC-Jabra] Evento: notificar_carro = %v", temCarro)
		// TODO: Tocar som, mostrar notificação, etc.
	})

	app.Socket.OnLigacaoAtendida(func(ramalQueAtendeu string) {
		log.Printf("[ACC-Jabra] Evento: ligacao_atendida por ramal %s", ramalQueAtendeu)
	})

	app.Socket.OnLigacaoInterna(func(ramalQueSolicitou string, temCarro bool) {
		log.Printf("[ACC-Jabra] Evento: ligacao_interna de %s (carro: %v)", ramalQueSolicitou, temCarro)
	})
}

func loadSocketConfig() SocketConfig {
	config := SocketConfig{}

	// Tenta carregar de arquivo
	configPath := getConfigPath("socket.json")
	data, err := os.ReadFile(configPath)
	if err == nil {
		json.Unmarshal(data, &config)
	}

	// Override com settings do banco
	if host := app.Store.GetSetting("socket_host", ""); host != "" {
		config.Host = host
	}
	if token := app.Store.GetSetting("token", ""); token != "" {
		config.Token = token
	}
	if ramal := app.Store.GetSetting("ramal", ""); ramal != "" {
		config.Ramal = ramal
	}

	return config
}

func getConfigPath(filename string) string {
	// Primeiro tenta no diretório config/ relativo ao executável
	execPath, _ := os.Executable()
	execDir := filepath.Dir(execPath)

	configPath := filepath.Join(execDir, "config", filename)
	if _, err := os.Stat(configPath); err == nil {
		return configPath
	}

	// Fallback para diretório atual
	configPath = filepath.Join("config", filename)
	if _, err := os.Stat(configPath); err == nil {
		return configPath
	}

	// Retorna path padrão mesmo se não existir
	return filepath.Join("config", filename)
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Ícone minimalista PNG (círculo azul 16x16)
var iconData = []byte{
	0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
	0x00, 0x00, 0x00, 0x10, 0x00, 0x00, 0x00, 0x10, 0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x91, 0x68,
	0x36, 0x00, 0x00, 0x00, 0x19, 0x49, 0x44, 0x41, 0x54, 0x78, 0xda, 0x63, 0x64, 0xf8, 0xff, 0xbf,
	0x1e, 0x03, 0x06, 0x18, 0xec, 0x19, 0x10, 0x12, 0x03, 0x20, 0x54, 0x00, 0x00, 0x2e, 0x0e, 0x01,
	0x01, 0x8b, 0x21, 0x4d, 0x25, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae, 0x42, 0x60,
	0x82,
}
