package actions

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/gen2brain/beeep"
)

// ActionType define os tipos de ação disponíveis
type ActionType string

const (
	ActionAPICall    ActionType = "api_call"    // Faz chamada HTTP GET/POST
	ActionExec       ActionType = "exec"        // Executa comando do sistema
	ActionSocketEmit ActionType = "socket_emit" // Emite evento via Socket.IO
	ActionNotify     ActionType = "notify"      // Mostra notificação do sistema
	ActionPlaySound  ActionType = "play_sound"  // Reproduz som
	ActionNone       ActionType = "none"        // Não faz nada
)

// Action define uma ação a ser executada quando um botão é pressionado
type Action struct {
	Type    ActionType `json:"action"`
	URL     string     `json:"url,omitempty"`     // Para api_call
	Method  string     `json:"method,omitempty"`  // Para api_call (GET, POST)
	Body    string     `json:"body,omitempty"`    // Para api_call POST
	Command string     `json:"cmd,omitempty"`     // Para exec
	Event   string     `json:"event,omitempty"`   // Para socket_emit
	Message string     `json:"message,omitempty"` // Para notify
	Title   string     `json:"title,omitempty"`   // Para notify
	Sound   string     `json:"sound,omitempty"`   // Para play_sound (path do arquivo)
}

// KeyMap mapeia IDs de botão para ações
type KeyMap map[string]Action

// SocketEmitter interface para emitir eventos Socket.IO
type SocketEmitter interface {
	EmitClick(button string) error
}

// Executor gerencia a execução de ações baseado em eventos de botão
type Executor struct {
	mu       sync.RWMutex
	keyMap   KeyMap
	filePath string
	socket   SocketEmitter

	// Debounce para evitar execuções duplicadas
	lastExecution map[string]time.Time
	debounceTime  time.Duration
}

// NewExecutor cria um novo executor de ações
func NewExecutor(keymapPath string) (*Executor, error) {
	e := &Executor{
		keyMap:        make(KeyMap),
		filePath:      keymapPath,
		lastExecution: make(map[string]time.Time),
		debounceTime:  200 * time.Millisecond,
	}

	if keymapPath != "" {
		if err := e.LoadKeyMap(keymapPath); err != nil {
			// Se arquivo não existe, usa keymap padrão
			if os.IsNotExist(err) {
				e.keyMap = DefaultKeyMap()
				log.Printf("[Actions] Usando keymap padrão, arquivo não encontrado: %s", keymapPath)
			} else {
				return nil, err
			}
		}
	} else {
		e.keyMap = DefaultKeyMap()
	}

	return e, nil
}

// DefaultKeyMap retorna o mapeamento padrão de botões
func DefaultKeyMap() KeyMap {
	return KeyMap{
		"OffHook": {
			Type:  ActionSocketEmit,
			Event: "click",
		},
		"Mute": {
			Type:    ActionNotify,
			Title:   "ACC Jabra",
			Message: "Mute ativado",
		},
		"VolumeUp": {
			Type:    ActionNone,
			Message: "Volume aumentado",
		},
		"VolumeDown": {
			Type:    ActionNone,
			Message: "Volume diminuído",
		},
		"HookSwitch": {
			Type:  ActionSocketEmit,
			Event: "click",
		},
	}
}

// LoadKeyMap carrega o mapeamento de um arquivo JSON
func (e *Executor) LoadKeyMap(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var km KeyMap
	if err := json.Unmarshal(data, &km); err != nil {
		return fmt.Errorf("invalid keymap JSON: %w", err)
	}

	e.mu.Lock()
	e.keyMap = km
	e.filePath = path
	e.mu.Unlock()

	log.Printf("[Actions] KeyMap carregado: %d mapeamentos", len(km))
	return nil
}

// SaveKeyMap salva o mapeamento atual para arquivo
func (e *Executor) SaveKeyMap(path string) error {
	e.mu.RLock()
	data, err := json.MarshalIndent(e.keyMap, "", "  ")
	e.mu.RUnlock()

	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// SetSocketEmitter define o cliente Socket.IO para emissão de eventos
func (e *Executor) SetSocketEmitter(socket SocketEmitter) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.socket = socket
}

// Execute executa a ação mapeada para um botão
func (e *Executor) Execute(buttonID string, pressed bool) error {
	// Só executa no press, não no release (exceto se configurado diferente)
	if !pressed {
		return nil
	}

	e.mu.RLock()
	action, ok := e.keyMap[buttonID]
	socket := e.socket
	e.mu.RUnlock()

	if !ok {
		log.Printf("[Actions] Botão não mapeado: %s", buttonID)
		return nil
	}

	// Debounce
	e.mu.Lock()
	lastTime, exists := e.lastExecution[buttonID]
	now := time.Now()
	if exists && now.Sub(lastTime) < e.debounceTime {
		e.mu.Unlock()
		return nil
	}
	e.lastExecution[buttonID] = now
	e.mu.Unlock()

	log.Printf("[Actions] Executando ação para botão %s: %s", buttonID, action.Type)

	switch action.Type {
	case ActionAPICall:
		return e.executeAPICall(action)
	case ActionExec:
		return e.executeCommand(action)
	case ActionSocketEmit:
		return e.executeSocketEmit(action, buttonID, socket)
	case ActionNotify:
		return e.executeNotify(action)
	case ActionPlaySound:
		return e.executePlaySound(action)
	case ActionNone:
		return nil
	default:
		return fmt.Errorf("unknown action type: %s", action.Type)
	}
}

// executeAPICall faz uma chamada HTTP
func (e *Executor) executeAPICall(action Action) error {
	method := action.Method
	if method == "" {
		method = "GET"
	}

	var resp *http.Response
	var err error

	switch strings.ToUpper(method) {
	case "GET":
		resp, err = http.Get(action.URL)
	case "POST":
		resp, err = http.Post(action.URL, "application/json", strings.NewReader(action.Body))
	default:
		return fmt.Errorf("unsupported HTTP method: %s", method)
	}

	if err != nil {
		return fmt.Errorf("API call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("API call returned status %d", resp.StatusCode)
	}

	log.Printf("[Actions] API call para %s retornou %d", action.URL, resp.StatusCode)
	return nil
}

// executeCommand executa um comando do sistema
func (e *Executor) executeCommand(action Action) error {
	if action.Command == "" {
		return fmt.Errorf("no command specified")
	}

	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", action.Command)
	} else {
		cmd = exec.Command("sh", "-c", action.Command)
	}

	// Executa em background
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("command start failed: %w", err)
	}

	log.Printf("[Actions] Comando iniciado: %s", action.Command)
	return nil
}

// executeSocketEmit emite evento via Socket.IO
func (e *Executor) executeSocketEmit(action Action, buttonID string, socket SocketEmitter) error {
	if socket == nil {
		return fmt.Errorf("socket emitter not configured")
	}

	event := action.Event
	if event == "" {
		event = buttonID
	}

	if err := socket.EmitClick(event); err != nil {
		return fmt.Errorf("socket emit failed: %w", err)
	}

	log.Printf("[Actions] Evento Socket.IO emitido: %s", event)
	return nil
}

// executeNotify mostra notificação do sistema
func (e *Executor) executeNotify(action Action) error {
	title := action.Title
	if title == "" {
		title = "ACC Jabra"
	}

	if err := beeep.Notify(title, action.Message, ""); err != nil {
		return fmt.Errorf("notification failed: %w", err)
	}

	log.Printf("[Actions] Notificação exibida: %s", action.Message)
	return nil
}

// executePlaySound reproduz um arquivo de som
func (e *Executor) executePlaySound(action Action) error {
	if action.Sound == "" {
		return fmt.Errorf("no sound file specified")
	}

	// Usa beeep.Beep para som simples, ou abre arquivo via comando do sistema
	if action.Sound == "beep" {
		return beeep.Beep(beeep.DefaultFreq, beeep.DefaultDuration)
	}

	// Reproduz arquivo de áudio via comando do sistema
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		// Windows: usa PowerShell para tocar som
		ps := fmt.Sprintf("(New-Object Media.SoundPlayer '%s').PlaySync()", action.Sound)
		cmd = exec.Command("powershell", "-Command", ps)
	} else {
		// Linux: tenta aplay ou paplay
		cmd = exec.Command("aplay", action.Sound)
	}

	return cmd.Start()
}

// GetKeyMap retorna o mapeamento atual
func (e *Executor) GetKeyMap() KeyMap {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Retorna cópia
	km := make(KeyMap, len(e.keyMap))
	for k, v := range e.keyMap {
		km[k] = v
	}
	return km
}

// SetAction define a ação para um botão específico
func (e *Executor) SetAction(buttonID string, action Action) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.keyMap[buttonID] = action
}

// RemoveAction remove a ação de um botão
func (e *Executor) RemoveAction(buttonID string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.keyMap, buttonID)
}
