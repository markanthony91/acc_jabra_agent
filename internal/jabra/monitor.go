package jabra

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/aiknow/acc_jabra_agent/internal/db"
	"github.com/aiknow/acc_jabra_agent/internal/models"
	"github.com/gen2brain/beeep"
	"github.com/karalabe/hid"
)

const JabraVID = 0x0b0e

type Monitor struct {
	currentState models.TelemetryPayload
	lastUpdate   time.Time
	mu           sync.RWMutex
	store        *db.Store
}

func NewMonitor(serial string, store *db.Store) *Monitor {
	m := &Monitor{
		store: store,
		currentState: models.TelemetryPayload{
			Module: "jabra_telemetry",
			Device: "Engage 55 Mono SE",
			Serial: serial,
			State: models.DeviceState{
				Battery:       models.BatteryInfo{Level: 0, Status: "unknown"},
				Connection:    "offline",
				SessionUptime: "00h 00m",
				CustomID:      "Aguardando...",
				CustomColor:   "#9e9e9e",
			},
			Events: models.DeviceEvents{LastPowerOn: time.Now()},
		},
		lastUpdate: time.Now(),
	}

	go m.startHIDScanner()
	go m.batteryLogger()
	go m.uptimeUpdater()
	return m
}

func (m *Monitor) uptimeUpdater() {
	ticker := time.NewTicker(30 * time.Second)
	for range ticker.C {
		m.mu.Lock()
		if m.currentState.State.Connection == "online" {
			duration := time.Since(m.currentState.Events.LastPowerOn)
			hours := int(duration.Hours())
			minutes := int(duration.Minutes()) % 60
			m.currentState.State.SessionUptime = fmt.Sprintf("%02dh %02dm", hours, minutes)
		}
		m.mu.Unlock()
	}
}

func (m *Monitor) setConnectionStatus(status string, deviceName string, serial string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.currentState.State.Connection != status {
		m.currentState.State.Connection = status
		if status == "online" {
			m.currentState.Device = deviceName
			m.currentState.Serial = serial
			m.currentState.Events.LastPowerOn = time.Now()
			
			// Carrega identidade das configurações
			m.currentState.State.CustomID = m.store.GetSetting("operator_name", "Operador 01")
			m.currentState.State.CustomColor = m.store.GetSetting("custom_color", "#2196F3")
			
			beeep.Notify("ACC Jabra", "Headset Conectado: "+deviceName, "")
		} else {
			m.currentState.State.SessionUptime = "00h 00m"
			m.currentState.State.CustomID = "Desconectado"
			m.currentState.State.CustomColor = "#9e9e9e"
			beeep.Alert("ACC Jabra: ALERTA", "Dongle Removido!", "")
		}
		m.store.LogEvent("connection_change", "Status: "+status)
	}
}

func (m *Monitor) startHIDScanner() {
	for {
		devices := hid.Enumerate(JabraVID, 0)
		if len(devices) > 0 {
			info := devices[0]
			serial := info.Serial
			if serial == "" {
				serial = "USB-HID-DEVICE"
			}

			device, err := info.Open()
			if err != nil {
				log.Printf("[Jabra-HID] Erro: %v", err)
				m.runSimulation()
			} else {
				m.setConnectionStatus("online", info.Product, serial)
				m.readHIDEvents(device)
				device.Close()
				m.setConnectionStatus("offline", "", "")
			}
		} else {
			// Se não houver hardware, rodamos em modo simulação para desenvolvimento
			m.runSimulation()
		}
		time.Sleep(2 * time.Second)
	}
}

func (m *Monitor) runSimulation() {
	m.mu.Lock()
	if m.currentState.State.Connection == "online" {
		m.mu.Unlock()
		return
	}
	m.mu.Unlock()

	m.setConnectionStatus("online", "Jabra Simulation Mode", "SIM-123456")
	
	// Simular variação de bateria
	go func() {
		level := 100
		for {
			m.mu.Lock()
			if m.currentState.Serial != "SIM-123456" {
				m.mu.Unlock()
				return
			}
			m.currentState.State.Battery.Level = level
			m.currentState.State.Battery.Status = "discharging"
			m.currentState.State.Battery.EstimatedRemainingMinutes = m.CalculateRemainingMinutes(level, 0.1)
			m.mu.Unlock()

			level -= 1
			if level < 0 {
				level = 100
			}
			time.Sleep(10 * time.Second)
		}
	}()
}

func (m *Monitor) readHIDEvents(device *hid.Device) {
	buf := make([]byte, 64)
	for {
		n, err := device.Read(buf)
		if err != nil {
			break
		}
		if n > 0 {
			m.processHIDData(buf[:n])
		}
	}
}

func (m *Monitor) processHIDData(data []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Log para debug de pacotes novos
	log.Printf("[Jabra-HID] Packet: %v", data)

	// Mute Toggle (Padrão observado em alguns headsets)
	if data[0] == 0x01 && data[1] == 0x02 {
		m.currentState.Events.LastButtonPressed = "mute_toggle"
		m.currentState.State.IsMuted = !m.currentState.State.IsMuted
		m.store.LogEvent("button", "Mute Toggled")
		return
	}

	// Hook Switch (Call answer/end)
	if data[0] == 0x01 && (data[1] == 0x04 || data[1] == 0x08) {
		m.currentState.Events.LastButtonPressed = "hook_switch"
		m.currentState.State.IsInCall = !m.currentState.State.IsInCall
		m.store.LogEvent("button", "Hook Switch Toggled")
		return
	}
}

func (m *Monitor) batteryLogger() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		m.mu.RLock()
		status := m.currentState.State.Connection
		level := m.currentState.State.Battery.Level
		m.mu.RUnlock()

		if status == "online" {
			m.store.LogBattery(level, "periodic_check")
		}
	}
}

func (m *Monitor) CalculateRemainingMinutes(currentLevel int, dischargeRate float64) int {
	if currentLevel <= 0 { return 0 }
	if dischargeRate <= 0 { return 540 }
	return int(float64(currentLevel) / dischargeRate)
}

func (m *Monitor) GetTelemetry() models.TelemetryPayload {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentState
}
