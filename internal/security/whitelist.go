package security

import (
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"
)

// BlockMode define como dispositivos não autorizados são tratados
type BlockMode string

const (
	// BlockModeNone não bloqueia dispositivos
	BlockModeNone BlockMode = "none"

	// BlockModeSoft força mute e desliga ringer
	BlockModeSoft BlockMode = "soft"

	// BlockModeNotify apenas notifica sobre dispositivo não autorizado
	BlockModeNotify BlockMode = "notify"
)

// DeviceController interface para controlar dispositivos
type DeviceController interface {
	SetMute(deviceID uint16, mute bool) error
	SetRinger(deviceID uint16, ring bool) error
}

// WhitelistConfig é a configuração do arquivo JSON
type WhitelistConfig struct {
	Enabled        bool      `json:"enabled"`
	BlockMode      BlockMode `json:"block_mode"`
	AllowedSerials []string  `json:"allowed_serials"`
}

// Whitelist gerencia a lista de dispositivos autorizados
type Whitelist struct {
	mu             sync.RWMutex
	config         WhitelistConfig
	filePath       string
	allowedSerials map[string]bool

	// Dispositivos atualmente bloqueados (soft-block ativo)
	blockedDevices map[uint16]bool

	// Intervalo para reforçar soft-block
	enforceInterval time.Duration
	stopEnforce     chan struct{}
}

// NewWhitelist cria uma nova instância do whitelist
func NewWhitelist(configPath string) (*Whitelist, error) {
	w := &Whitelist{
		config: WhitelistConfig{
			Enabled:   false,
			BlockMode: BlockModeNone,
		},
		filePath:        configPath,
		allowedSerials:  make(map[string]bool),
		blockedDevices:  make(map[uint16]bool),
		enforceInterval: 5 * time.Second,
		stopEnforce:     make(chan struct{}),
	}

	if configPath != "" {
		if err := w.LoadConfig(configPath); err != nil {
			if !os.IsNotExist(err) {
				return nil, err
			}
			// Arquivo não existe, usa configuração padrão
			log.Printf("[Whitelist] Arquivo não encontrado, whitelist desabilitado: %s", configPath)
		}
	}

	return w, nil
}

// LoadConfig carrega a configuração do arquivo JSON
func (w *Whitelist) LoadConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var config WhitelistConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}

	w.mu.Lock()
	w.config = config
	w.filePath = path

	// Converte lista para map para lookup O(1)
	w.allowedSerials = make(map[string]bool, len(config.AllowedSerials))
	for _, serial := range config.AllowedSerials {
		w.allowedSerials[serial] = true
	}
	w.mu.Unlock()

	log.Printf("[Whitelist] Configuração carregada: %d seriais permitidos, modo: %s",
		len(config.AllowedSerials), config.BlockMode)

	return nil
}

// SaveConfig salva a configuração atual para arquivo
func (w *Whitelist) SaveConfig(path string) error {
	w.mu.RLock()
	data, err := json.MarshalIndent(w.config, "", "  ")
	w.mu.RUnlock()

	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// IsEnabled retorna se o whitelist está habilitado
func (w *Whitelist) IsEnabled() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.config.Enabled
}

// SetEnabled habilita ou desabilita o whitelist
func (w *Whitelist) SetEnabled(enabled bool) {
	w.mu.Lock()
	w.config.Enabled = enabled
	w.mu.Unlock()
}

// IsAllowed verifica se um serial está na whitelist
func (w *Whitelist) IsAllowed(serial string) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()

	// Se whitelist desabilitado, permite todos
	if !w.config.Enabled {
		return true
	}

	// Se whitelist vazio, permite todos
	if len(w.allowedSerials) == 0 {
		return true
	}

	return w.allowedSerials[serial]
}

// AddSerial adiciona um serial à whitelist
func (w *Whitelist) AddSerial(serial string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.allowedSerials[serial] = true

	// Atualiza config
	found := false
	for _, s := range w.config.AllowedSerials {
		if s == serial {
			found = true
			break
		}
	}
	if !found {
		w.config.AllowedSerials = append(w.config.AllowedSerials, serial)
	}
}

// RemoveSerial remove um serial da whitelist
func (w *Whitelist) RemoveSerial(serial string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	delete(w.allowedSerials, serial)

	// Atualiza config
	newList := make([]string, 0, len(w.config.AllowedSerials))
	for _, s := range w.config.AllowedSerials {
		if s != serial {
			newList = append(newList, s)
		}
	}
	w.config.AllowedSerials = newList
}

// GetAllowedSerials retorna lista de seriais permitidos
func (w *Whitelist) GetAllowedSerials() []string {
	w.mu.RLock()
	defer w.mu.RUnlock()

	serials := make([]string, 0, len(w.config.AllowedSerials))
	serials = append(serials, w.config.AllowedSerials...)
	return serials
}

// GetBlockMode retorna o modo de bloqueio atual
func (w *Whitelist) GetBlockMode() BlockMode {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.config.BlockMode
}

// SetBlockMode define o modo de bloqueio
func (w *Whitelist) SetBlockMode(mode BlockMode) {
	w.mu.Lock()
	w.config.BlockMode = mode
	w.mu.Unlock()
}

// CheckDevice verifica um dispositivo e retorna se deve ser bloqueado
// Retorna: allowed, shouldBlock
func (w *Whitelist) CheckDevice(serial string) (allowed bool, shouldBlock bool) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if !w.config.Enabled {
		return true, false
	}

	allowed = w.allowedSerials[serial] || len(w.allowedSerials) == 0
	shouldBlock = !allowed && w.config.BlockMode == BlockModeSoft

	return
}

// SoftBlock aplica soft-block em um dispositivo não autorizado
func (w *Whitelist) SoftBlock(controller DeviceController, deviceID uint16) error {
	w.mu.Lock()
	w.blockedDevices[deviceID] = true
	w.mu.Unlock()

	// Força mute e desliga ringer
	if err := controller.SetMute(deviceID, true); err != nil {
		log.Printf("[Whitelist] Erro ao aplicar mute no dispositivo %d: %v", deviceID, err)
	}

	if err := controller.SetRinger(deviceID, false); err != nil {
		log.Printf("[Whitelist] Erro ao desligar ringer no dispositivo %d: %v", deviceID, err)
	}

	log.Printf("[Whitelist] Soft-block aplicado no dispositivo %d", deviceID)
	return nil
}

// RemoveBlock remove o bloqueio de um dispositivo
func (w *Whitelist) RemoveBlock(deviceID uint16) {
	w.mu.Lock()
	delete(w.blockedDevices, deviceID)
	w.mu.Unlock()
}

// IsBlocked verifica se um dispositivo está bloqueado
func (w *Whitelist) IsBlocked(deviceID uint16) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.blockedDevices[deviceID]
}

// StartEnforcement inicia goroutine que reforça soft-block periodicamente
func (w *Whitelist) StartEnforcement(controller DeviceController) {
	go func() {
		ticker := time.NewTicker(w.enforceInterval)
		defer ticker.Stop()

		for {
			select {
			case <-w.stopEnforce:
				return
			case <-ticker.C:
				w.enforceBlocks(controller)
			}
		}
	}()
}

// StopEnforcement para a goroutine de enforcement
func (w *Whitelist) StopEnforcement() {
	close(w.stopEnforce)
	w.stopEnforce = make(chan struct{})
}

// enforceBlocks reaplica soft-block em dispositivos bloqueados
func (w *Whitelist) enforceBlocks(controller DeviceController) {
	w.mu.RLock()
	devices := make([]uint16, 0, len(w.blockedDevices))
	for id := range w.blockedDevices {
		devices = append(devices, id)
	}
	w.mu.RUnlock()

	for _, deviceID := range devices {
		controller.SetMute(deviceID, true)
	}
}

// GetConfig retorna cópia da configuração atual
func (w *Whitelist) GetConfig() WhitelistConfig {
	w.mu.RLock()
	defer w.mu.RUnlock()

	config := w.config
	config.AllowedSerials = make([]string, len(w.config.AllowedSerials))
	copy(config.AllowedSerials, w.config.AllowedSerials)
	return config
}
