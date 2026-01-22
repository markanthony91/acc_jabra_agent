# ACC Jabra Telemetry Agent

Agente de telemetria de hardware escrito em **Go** para monitoramento avançado de headsets **Jabra Engage 55 Mono SE**. Parte integrante do ecossistema Aiknow Command Center (ACC).

## 🚀 Funcionalidades

- **Monitoramento em Tempo Real:** Status de chamada, mudo e volume.
- **Session Tracking:** Cronômetro de Uptime (tempo logado) por sessão.
- **Identidade Visual:** Suporte a Custom ID (Operador) e Cor de identificação.
- **Telemetria de Bateria:** Nível atual e status de carregamento.
- **Predictive Analytics:** Cálculo estimado de minutos restantes de conversação.
- **Event Tracking:** Captura de eventos de botões físicos via USB HID.
- **API REST:** Endpoint JSON para integração com ACC Core e Dashboards.

## 🛠 Tecnologia

- **Linguagem:** Go 1.22+
- **Integração:** USB HID (Vendor ID: `0b0e`)
- **Ambiente:** Nix Flakes para desenvolvimento reprodutível
- **Arquitetura:** Concorrência via Goroutines para scanner USB não bloqueante.

## 📦 Instalação e Uso

### Requisitos
- Linux (Ubuntu/Fedora) ou WSL2
- Nix (recomendado) ou Go instalado

### Executando com Nix
```bash
nix develop
go mod tidy
go run cmd/agent/main.go
```

### Endpoints da API
- `GET /api/telemetry`: Retorna o estado completo do dispositivo e telemetria.
- `GET /api/health`: Health check do agente.

## 📊 Estrutura de Resposta
```json
{
  "hostname": "workstation-01",
  "data": {
    "module": "jabra_telemetry",
    "device": "Engage 55 Mono SE",
    "serial": "ABC123456789",
    "state": {
      "is_in_call": false,
      "is_muted": false,
      "volume": 75,
      "battery": {
        "level": 82,
        "status": "discharging",
        "estimated_remaining_minutes": 540
      }
    }
  }
}
```

---
*Desenvolvido para Aiknow Systems - Marcelo.*
