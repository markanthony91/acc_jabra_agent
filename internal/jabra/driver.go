package jabra

import "time"

// ButtonID representa os IDs de botões traduzidos do Jabra SDK
type ButtonID int

const (
	ButtonCyclic ButtonID = iota
	ButtonCyclicEnd
	ButtonDecline
	ButtonDialNext
	ButtonDialPrev
	ButtonEndCall
	ButtonFireAlarm
	ButtonFlash
	ButtonFlexibleBootMute
	ButtonGN1
	ButtonGN2
	ButtonGN3
	ButtonGN4
	ButtonGN5
	ButtonGN6
	ButtonHookSwitch
	ButtonJabra
	ButtonKey0
	ButtonKey1
	ButtonKey2
	ButtonKey3
	ButtonKey4
	ButtonKey5
	ButtonKey6
	ButtonKey7
	ButtonKey8
	ButtonKey9
	ButtonKeyClear
	ButtonKeyPound
	ButtonKeyStar
	ButtonLineBusy
	ButtonMute
	ButtonOffline
	ButtonOffHook
	ButtonOnline
	ButtonPseudoOffHook
	ButtonRedial
	ButtonRejectCall
	ButtonSpeedDial
	ButtonTransfer
	ButtonVoiceMail
	ButtonVolumeDown
	ButtonVolumeUp
)

// String retorna o nome do botão para logging e keymap
func (b ButtonID) String() string {
	names := map[ButtonID]string{
		ButtonCyclic:           "Cyclic",
		ButtonCyclicEnd:        "CyclicEnd",
		ButtonDecline:          "Decline",
		ButtonDialNext:         "DialNext",
		ButtonDialPrev:         "DialPrev",
		ButtonEndCall:          "EndCall",
		ButtonFireAlarm:        "FireAlarm",
		ButtonFlash:            "Flash",
		ButtonFlexibleBootMute: "FlexibleBootMute",
		ButtonGN1:              "GN1",
		ButtonGN2:              "GN2",
		ButtonGN3:              "GN3",
		ButtonGN4:              "GN4",
		ButtonGN5:              "GN5",
		ButtonGN6:              "GN6",
		ButtonHookSwitch:       "HookSwitch",
		ButtonJabra:            "Jabra",
		ButtonKey0:             "Key0",
		ButtonKey1:             "Key1",
		ButtonKey2:             "Key2",
		ButtonKey3:             "Key3",
		ButtonKey4:             "Key4",
		ButtonKey5:             "Key5",
		ButtonKey6:             "Key6",
		ButtonKey7:             "Key7",
		ButtonKey8:             "Key8",
		ButtonKey9:             "Key9",
		ButtonKeyClear:         "KeyClear",
		ButtonKeyPound:         "KeyPound",
		ButtonKeyStar:          "KeyStar",
		ButtonLineBusy:         "LineBusy",
		ButtonMute:             "Mute",
		ButtonOffline:          "Offline",
		ButtonOffHook:          "OffHook",
		ButtonOnline:           "Online",
		ButtonPseudoOffHook:    "PseudoOffHook",
		ButtonRedial:           "Redial",
		ButtonRejectCall:       "RejectCall",
		ButtonSpeedDial:        "SpeedDial",
		ButtonTransfer:         "Transfer",
		ButtonVoiceMail:        "VoiceMail",
		ButtonVolumeDown:       "VolumeDown",
		ButtonVolumeUp:         "VolumeUp",
	}
	if name, ok := names[b]; ok {
		return name
	}
	return "Unknown"
}

// DeviceInfo contém informações sobre um dispositivo Jabra
type DeviceInfo struct {
	ID           uint16
	Name         string
	SerialNumber string
	VendorID     uint16
	ProductID    uint16
	IsDongle     bool
	Connected    bool
	ConnectedAt  time.Time
}

// BatteryStatus contém informações sobre a bateria do dispositivo
type BatteryStatus struct {
	Level      int  // 0-100
	IsCharging bool
	IsLow      bool
}

// ButtonEvent representa um evento de botão
type ButtonEvent struct {
	DeviceID uint16
	ButtonID ButtonID
	Pressed  bool // true = pressionado, false = liberado
}

// DeviceEvent representa um evento de dispositivo (conectado/desconectado)
type DeviceEvent struct {
	DeviceID  uint16
	Connected bool
	Device    *DeviceInfo
}

// Driver é a interface abstrata para comunicação com dispositivos Jabra.
// Permite implementações diferentes para Windows (SDK) e Linux (HID).
type Driver interface {
	// Start inicia o driver e começa a monitorar dispositivos
	Start() error

	// Stop para o driver e libera recursos
	Stop() error

	// IsRunning retorna true se o driver está ativo
	IsRunning() bool

	// GetDevices retorna lista de dispositivos conectados
	GetDevices() []DeviceInfo

	// GetDevice retorna informações de um dispositivo específico
	GetDevice(deviceID uint16) (*DeviceInfo, error)

	// GetBatteryStatus obtém status da bateria de um dispositivo
	GetBatteryStatus(deviceID uint16) (*BatteryStatus, error)

	// SetMute define estado do mute
	SetMute(deviceID uint16, mute bool) error

	// GetMute obtém estado do mute
	GetMute(deviceID uint16) (bool, error)

	// SetRinger define estado do ringer (toque)
	SetRinger(deviceID uint16, ring bool) error

	// SetHookState define estado do hook (off-hook = atendendo)
	SetHookState(deviceID uint16, offHook bool) error

	// SetBusylight define estado do LED de ocupado
	SetBusylight(deviceID uint16, on bool) error

	// SetHold define estado de hold (chamada em espera)
	SetHold(deviceID uint16, hold bool) error

	// SetVolume define volume do dispositivo (0-100)
	SetVolume(deviceID uint16, volume int) error

	// GetVolume obtém volume do dispositivo
	GetVolume(deviceID uint16) (int, error)

	// OnDeviceConnected registra callback para dispositivo conectado
	OnDeviceConnected(handler func(event DeviceEvent))

	// OnDeviceDisconnected registra callback para dispositivo desconectado
	OnDeviceDisconnected(handler func(event DeviceEvent))

	// OnButtonEvent registra callback para eventos de botão
	OnButtonEvent(handler func(event ButtonEvent))

	// OnBatteryUpdate registra callback para atualização de bateria
	OnBatteryUpdate(handler func(deviceID uint16, status BatteryStatus))
}

// DriverConfig contém configurações para inicialização do driver
type DriverConfig struct {
	// AppID é o ID da aplicação para o Jabra SDK (Windows)
	AppID string

	// VendorID para filtrar dispositivos (HID)
	VendorID uint16

	// PollInterval é o intervalo de polling para HID driver
	PollInterval time.Duration

	// SimulationMode ativa modo de simulação quando hardware não está disponível
	SimulationMode bool
}

// DefaultConfig retorna configuração padrão
func DefaultConfig() DriverConfig {
	return DriverConfig{
		AppID:          "88b7-5cbde35c-e588-49b3-a6d5-f54278270e28",
		VendorID:       0x0b0e, // Jabra Vendor ID
		PollInterval:   2 * time.Second,
		SimulationMode: false,
	}
}
