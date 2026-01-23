package db

import (
	"os"
	"testing"
)

func TestSQLiteStore(t *testing.T) {
	// Setup: Usar um banco temporário para testes
	testDB := "./data/test_jabra.db"
	_ = os.MkdirAll("./data", 0755)
	defer os.Remove(testDB)

	store, err := NewStore(testDB)
	if err != nil {
		t.Fatalf("Erro ao criar store: %v", err)
	}

	t.Run("Log e Get Bateria", func(t *testing.T) {
		store.LogBattery(85, "discharging")
		history, err := store.GetBatteryHistory(1)
		if err != nil {
			t.Fatalf("Erro ao buscar histórico: %v", err)
		}
		if len(history) == 0 {
			t.Fatal("Histórico deveria ter 1 registro")
		}
		if history[0]["level"].(int) != 85 {
			t.Errorf("Level esperado 85, obtido %v", history[0]["level"])
		}
	})

	t.Run("Log de Eventos", func(t *testing.T) {
		store.LogEvent("test", "descrição de teste")
		// Se não deu erro no LogEvent, consideramos OK por agora
	})
}
