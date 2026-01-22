//go:build !windows

package jabra

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/karalabe/hid"
)

const (
	// JabraVendorID é o Vendor ID USB da Jabra
	JabraVendorID uint16 = 0x0b0e
)

// HIDDriver implementa Driver usando comunicação HID genérica (Linux/macOS)
type HIDDriver struct {
	mu      sync.RWMutex
	config  DriverConfig
	running bool
	stopCh  chan struct{}

	// Dispositivos conectados
	devices map[uint16]*DeviceInfo

	// Device HID atualmente aberto
	currentDevice *hid.Device

	// Callbacks
	onDeviceConnected    func(event DeviceEvent)
	onDeviceDisconnected func(event DeviceEvent)
	onButtonEvent        func(event ButtonEvent)
	onBatteryUpdate      func(deviceID uint16, status BatteryStatus)

	// Estado simulado (quando hardware não disponível)
	simulationMode  bool
	simulatedBattery int
}

// NewHIDDriver cria uma nova instância do driver HID
func NewHIDDriver(config DriverConfig) (*HIDDriver, error) {
	return &HIDDriver{
		config:           config,
		devices:          make(map[uint16]*DeviceInfo),
		stopCh:           make(chan struct{}),
		simulatedBattery: 100,
	}, nil
}

// Start inicia o driver e começa a monitorar dispositivos
func (d *HIDDriver) Start() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.running {
		return errors.New("driver already running")
	}

	d.running = true
	d.stopCh = make(chan struct{})

	// Inicia scanner de dispositivos
	go d.scanLoop()

	log.Println("[HID Driver] Iniciado")
	return nil
}

// Stop para o driver
func (d *HIDDriver) Stop() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.running {
		return nil
	}

	close(d.stopCh)
	d.running = false

	// Fecha dispositivo atual
	if d.currentDevice != nil {
		d.currentDevice.Close()
		d.currentDevice = nil
	}

	log.Println("[HID Driver] Parado")
	return nil
}

// IsRunning retorna se o driver está ativo
func (d *HIDDriver) IsRunning() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.running
}

// scanLoop procura dispositivos Jabra periodicamente
func (d *HIDDriver) scanLoop() {
	ticker := time.NewTicker(d.config.PollInterval)
	defer ticker.Stop()

	// Scan inicial
	d.scanDevices()

	for {
		select {
		case <-d.stopCh:
			return
		case <-ticker.C:
			d.scanDevices()
		}
	}
}

// scanDevices enumera dispositivos HID Jabra
func (d *HIDDriver) scanDevices() {
	devices := hid.Enumerate(JabraVendorID, 0)

	d.mu.Lock()
	defer d.mu.Unlock()

	// Marca dispositivos conhecidos como potencialmente desconectados
	currentDeviceIDs := make(map[uint16]bool)

	for _, devInfo := range devices {
		deviceID := uint16(devInfo.ProductID)
		currentDeviceIDs[deviceID] = true

		// Dispositivo novo?
		if _, exists := d.devices[deviceID]; !exists {
			info := &DeviceInfo{
				ID:          deviceID,
				Name:        devInfo.Product,
				SerialNumber: devInfo.Serial,
				VendorID:    uint16(devInfo.VendorID),
				ProductID:   uint16(devInfo.ProductID),
				IsDongle:    d.isDongle(devInfo.Product),
				Connected:   true,
				ConnectedAt: time.Now(),
			}

			d.devices[deviceID] = info
			d.simulationMode = false

			log.Printf("[HID Driver] Dispositivo conectado: %s (ID: %d)", info.Name, info.ID)

			// Notifica callback
			if d.onDeviceConnected != nil {
				go d.onDeviceConnected(DeviceEvent{
					DeviceID:  deviceID,
					Connected: true,
					Device:    info,
				})
			}

			// Tenta abrir dispositivo para leitura de eventos
			go d.tryOpenDevice(devInfo)
		}
	}

	// Verifica dispositivos desconectados
	for id, info := range d.devices {
		if !currentDeviceIDs[id] && info.Connected {
			info.Connected = false

			log.Printf("[HID Driver] Dispositivo desconectado: %s (ID: %d)", info.Name, info.ID)

			if d.onDeviceDisconnected != nil {
				go d.onDeviceDisconnected(DeviceEvent{
					DeviceID:  id,
					Connected: false,
					Device:    info,
				})
			}

			delete(d.devices, id)
		}
	}

	// Ativa modo simulação se nenhum dispositivo
	if len(d.devices) == 0 && d.config.SimulationMode && !d.simulationMode {
		d.simulationMode = true
		log.Println("[HID Driver] Modo simulação ativado")
		go d.simulationLoop()
	}
}

// tryOpenDevice tenta abrir um dispositivo para leitura de eventos
func (d *HIDDriver) tryOpenDevice(devInfo hid.DeviceInfo) {
	device, err := devInfo.Open()
	if err != nil {
		log.Printf("[HID Driver] Não foi possível abrir dispositivo: %v", err)
		return
	}

	d.mu.Lock()
	d.currentDevice = device
	d.mu.Unlock()

	// Inicia leitura de eventos HID
	go d.readHIDEvents(device, uint16(devInfo.ProductID))
}

// readHIDEvents lê eventos HID do dispositivo
func (d *HIDDriver) readHIDEvents(device *hid.Device, deviceID uint16) {
	buf := make([]byte, 64)

	for {
		d.mu.RLock()
		running := d.running
		d.mu.RUnlock()

		if !running {
			return
		}

		n, err := device.Read(buf)
		if err != nil {
			log.Printf("[HID Driver] Erro ao ler HID: %v", err)
			return
		}

		if n > 0 {
			d.processHIDData(deviceID, buf[:n])
		}
	}
}

// processHIDData processa dados HID brutos e emite eventos de botão
func (d *HIDDriver) processHIDData(deviceID uint16, data []byte) {
	if len(data) < 2 {
		return
	}

	// Protocolo HID Jabra simplificado
	// Byte 0: Report ID
	// Byte 1: Estado dos botões

	reportID := data[0]
	buttonState := data[1]

	var buttonID ButtonID
	var pressed bool

	switch reportID {
	case 0x01: // Telephony page
		switch buttonState {
		case 0x01: // Hook off
			buttonID = ButtonOffHook
			pressed = true
		case 0x02: // Mute
			buttonID = ButtonMute
			pressed = true
		case 0x04: // Hook on
			buttonID = ButtonHookSwitch
			pressed = true
		case 0x08: // Flash
			buttonID = ButtonFlash
			pressed = true
		default:
			return
		}
	default:
		return
	}

	d.mu.RLock()
	handler := d.onButtonEvent
	d.mu.RUnlock()

	if handler != nil {
		handler(ButtonEvent{
			DeviceID: deviceID,
			ButtonID: buttonID,
			Pressed:  pressed,
		})
	}
}

// simulationLoop simula eventos quando não há hardware
func (d *HIDDriver) simulationLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-d.stopCh:
			return
		case <-ticker.C:
			d.mu.RLock()
			simMode := d.simulationMode
			d.mu.RUnlock()

			if !simMode {
				return
			}

			// Simula descarga de bateria
			d.mu.Lock()
			d.simulatedBattery -= 1
			if d.simulatedBattery < 0 {
				d.simulatedBattery = 100
			}
			battery := d.simulatedBattery
			d.mu.Unlock()

			if d.onBatteryUpdate != nil {
				d.onBatteryUpdate(0, BatteryStatus{
					Level:      battery,
					IsCharging: false,
					IsLow:      battery < 20,
				})
			}
		}
	}
}

// isDongle verifica se o dispositivo é um dongle baseado no nome
func (d *HIDDriver) isDongle(name string) bool {
	// Dongles comuns: Link 370, Link 380, etc.
	return len(name) > 0 && (name[0] == 'L' || name[0] == 'l')
}

// GetDevices retorna lista de dispositivos conectados
func (d *HIDDriver) GetDevices() []DeviceInfo {
	d.mu.RLock()
	defer d.mu.RUnlock()

	devices := make([]DeviceInfo, 0, len(d.devices))
	for _, dev := range d.devices {
		devices = append(devices, *dev)
	}

	// Adiciona dispositivo simulado se em modo simulação
	if d.simulationMode {
		devices = append(devices, DeviceInfo{
			ID:          0,
			Name:        "Jabra Simulado",
			Connected:   true,
			ConnectedAt: time.Now(),
		})
	}

	return devices
}

// GetDevice retorna informações de um dispositivo específico
func (d *HIDDriver) GetDevice(deviceID uint16) (*DeviceInfo, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.simulationMode && deviceID == 0 {
		return &DeviceInfo{
			ID:        0,
			Name:      "Jabra Simulado",
			Connected: true,
		}, nil
	}

	dev, ok := d.devices[deviceID]
	if !ok {
		return nil, fmt.Errorf("device %d not found", deviceID)
	}
	return dev, nil
}

// GetBatteryStatus obtém status da bateria
func (d *HIDDriver) GetBatteryStatus(deviceID uint16) (*BatteryStatus, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.simulationMode {
		return &BatteryStatus{
			Level:      d.simulatedBattery,
			IsCharging: false,
			IsLow:      d.simulatedBattery < 20,
		}, nil
	}

	// HID genérico não suporta leitura de bateria diretamente
	// Retorna valor padrão
	return &BatteryStatus{
		Level:      -1, // Desconhecido
		IsCharging: false,
		IsLow:      false,
	}, nil
}

// SetMute - HID genérico não suporta controle de mute
func (d *HIDDriver) SetMute(deviceID uint16, mute bool) error {
	log.Printf("[HID Driver] SetMute não suportado (mute=%v)", mute)
	return nil
}

// GetMute - HID genérico não suporta leitura de mute
func (d *HIDDriver) GetMute(deviceID uint16) (bool, error) {
	return false, nil
}

// SetRinger - HID genérico não suporta controle de ringer
func (d *HIDDriver) SetRinger(deviceID uint16, ring bool) error {
	log.Printf("[HID Driver] SetRinger não suportado (ring=%v)", ring)
	return nil
}

// SetHookState - HID genérico não suporta controle de hook
func (d *HIDDriver) SetHookState(deviceID uint16, offHook bool) error {
	log.Printf("[HID Driver] SetHookState não suportado (offHook=%v)", offHook)
	return nil
}

// SetBusylight - HID genérico não suporta controle de LED
func (d *HIDDriver) SetBusylight(deviceID uint16, on bool) error {
	log.Printf("[HID Driver] SetBusylight não suportado (on=%v)", on)
	return nil
}

// SetHold - HID genérico não suporta controle de hold
func (d *HIDDriver) SetHold(deviceID uint16, hold bool) error {
	log.Printf("[HID Driver] SetHold não suportado (hold=%v)", hold)
	return nil
}

// SetVolume - HID genérico não suporta controle de volume
func (d *HIDDriver) SetVolume(deviceID uint16, volume int) error {
	log.Printf("[HID Driver] SetVolume não suportado (volume=%d)", volume)
	return nil
}

// GetVolume - HID genérico não suporta leitura de volume
func (d *HIDDriver) GetVolume(deviceID uint16) (int, error) {
	return -1, nil
}

// OnDeviceConnected registra callback
func (d *HIDDriver) OnDeviceConnected(handler func(event DeviceEvent)) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.onDeviceConnected = handler
}

// OnDeviceDisconnected registra callback
func (d *HIDDriver) OnDeviceDisconnected(handler func(event DeviceEvent)) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.onDeviceDisconnected = handler
}

// OnButtonEvent registra callback
func (d *HIDDriver) OnButtonEvent(handler func(event ButtonEvent)) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.onButtonEvent = handler
}

// OnBatteryUpdate registra callback
func (d *HIDDriver) OnBatteryUpdate(handler func(deviceID uint16, status BatteryStatus)) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.onBatteryUpdate = handler
}
