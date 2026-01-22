package models

import "time"

type BatteryInfo struct {
	Level                     int    `json:"level"`
	Status                    string `json:"status"` // discharging, charging, fully charged
	EstimatedRemainingMinutes int    `json:"estimated_remaining_minutes"`
}

type DeviceState struct {
	IsInCall   bool        `json:"is_in_call"`
	IsMuted    bool        `json:"is_muted"`
	Volume     int         `json:"volume"`
	Battery    BatteryInfo `json:"battery"`
	Connection string      `json:"connection"` // stable, weak, disconnected
}

type DeviceEvents struct {
	LastPowerOn       time.Time `json:"last_power_on"`
	LastButtonPressed string    `json:"last_button_pressed"`
}

type TelemetryPayload struct {
	Module string       `json:"module"`
	Device string       `json:"device"`
	Serial string       `json:"serial"`
	State  DeviceState  `json:"state"`
	Events DeviceEvents `json:"events"`
}
