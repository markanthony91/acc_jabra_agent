package jabra

import (
	"testing"
)

func TestCalculateRemainingMinutes(t *testing.T) {
	m := &Monitor{}

	tests := []struct {
		name          string
		level         int
		rate          float64
		expected      int
	}{
		{"Bateria Cheia", 100, 0.1, 1000},
		{"Bateria Metade", 50, 0.2, 250},
		{"Bateria Baixa", 10, 0.5, 20},
		{"Rate Zero (Fallback)", 50, 0, 540},
		{"Bateria Zerada", 0, 0.1, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := m.CalculateRemainingMinutes(tt.level, tt.rate)
			if got != tt.expected {
				t.Errorf("CalculateRemainingMinutes() = %v, want %v", got, tt.expected)
			}
		})
	}
}
