package jabra

import (
	"log"
	"math"
	"sync"
	"time"

	"github.com/aiknow/acc_jabra_agent/internal/models"
	"github.com/karalabe/hid"
)

const (
	JabraVID = 0x0b0e // Jabra Vendor ID
)

type Monitor struct {
	currentState models.TelemetryPayload
	lastUpdate   time.Time
	mu           sync.RWMutex
}

func NewMonitor(serial string) *Monitor {
	m := &Monitor{
		currentState: models.TelemetryPayload{
			Module: "jabra_telemetry",
			Device: "Engage 55 Mono SE",
			Serial: serial,
			State: models.DeviceState{
				Battery: models.BatteryInfo{
					Level:  100,
					Status: "fully charged",
				},
				Connection: "stable",
			},
			Events: models.DeviceEvents{
				LastPowerOn: time.Now(),
			},
		},
		lastUpdate: time.Now(),
	}

	go m.startHIDScanner()
	return m
}

func (m *Monitor) startHIDScanner() {
	for {
		devices := hid.Enumerate(JabraVID, 0)
		if len(devices) > 0 {
			device, err := devices[0].Open()
			if err != nil {
				log.Printf("[Jabra-HID] Erro ao abrir dispositivo: %v", err)
			} else {
				log.Printf("[Jabra-HID] Dispositivo conectado: %s", devices[0].Product)
				m.readHIDEvents(device)
				device.Close()
			}
		}
		time.Sleep(5 * time.Second) // Tenta reconectar a cada 5s
	}
}

func (m *Monitor) readHIDEvents(device *hid.Device) {
	buf := make([]byte, 64)
	for {
		n, err := device.Read(buf)
		if err != nil {
			log.Printf("[Jabra-HID] Conexão perdida: %v", err)
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

	// Lógica simplificada de detecção de botões baseada em HID Reports comuns
	// Em uma implementação real, mapearíamos cada byte conforme o manual técnico do Engage 55
	m.currentState.Events.LastButtonPressed = "button_event_detected"
	log.Printf("[Jabra-HID] Evento capturado: %x", data)
}

func (m *Monitor) CalculateRemainingMinutes(currentLevel int, dischargeRate float64) int {
	if currentLevel <= 0 {
		return 0
	}
	if dischargeRate <= 0 {
		return 540
	}
	remaining := float64(currentLevel) / dischargeRate
	return int(math.Floor(remaining))
}

func (m *Monitor) GetTelemetry() models.TelemetryPayload {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentState
}