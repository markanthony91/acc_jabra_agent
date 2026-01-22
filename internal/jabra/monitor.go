package jabra

import (
	"log"
	"math"
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
				Battery:    models.BatteryInfo{Level: 0, Status: "unknown"},
				Connection: "offline",
			},
			Events: models.DeviceEvents{LastPowerOn: time.Now()},
		},
		lastUpdate: time.Now(),
	}

	go m.startHIDScanner()
	go m.batteryLogger()
	return m
}

func (m *Monitor) setConnectionStatus(status string, deviceName string, serial string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.currentState.State.Connection != status {
		m.currentState.State.Connection = status
		if status == "online" {
			m.currentState.Device = deviceName
			m.currentState.Serial = serial
			beeep.Notify("ACC Jabra", "Headset Conectado: "+deviceName, "")
		} else {
			beeep.Alert("ACC Jabra: ALERTA", "Dongle Removido ou Headset Desconectado!", "")
		}
		m.store.LogEvent("connection_change", "Status: "+status)
	}
}

func (m *Monitor) startHIDScanner() {
	for {
		devices := hid.Enumerate(JabraVID, 0)
		if len(devices) > 0 {
			// Tenta capturar Serial e Nome real
			info := devices[0]
			serial := info.Serial
			if serial == "" { serial = "USB-HID-DEVICE" }

			device, err := info.Open()
			if err != nil {
				log.Printf("[Jabra-HID] Erro: %v", err)
			} else {
				m.setConnectionStatus("online", info.Product, serial)
				m.readHIDEvents(device)
				device.Close()
				m.setConnectionStatus("offline", "", "")
			}
		} else {
			m.setConnectionStatus("offline", "", "")
		}
		time.Sleep(2 * time.Second) // Scanner mais rápido para detecção de dongle
	}
}

func (m *Monitor) readHIDEvents(device *hid.Device) {
	buf := make([]byte, 64)
	for {
		n, err := device.Read(buf)
		if err != nil { break }
		if n > 0 { m.processHIDData(buf[:n]) }
	}
}

func (m *Monitor) processHIDData(data []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Detecção de Botões (Engage 55)
	if data[0] == 0x01 && data[1] == 0x02 {
		m.currentState.Events.LastButtonPressed = "mute_toggle"
		m.currentState.State.IsMuted = !m.currentState.State.IsMuted
		m.store.LogEvent("button", "Mute Toggled")
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
			if level < 15 && level > 0 {
				beeep.Notify("ACC Jabra: Bateria Baixa", "Carga atual: ", "")
			}
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