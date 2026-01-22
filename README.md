# ACC Jabra Telemetry Agent

Agente de telemetria de hardware escrito em **Go** para monitoramento avançado de headsets Jabra. Parte integrante do ecossistema Aiknow Command Center (ACC), focado em fornecer telemetria em tempo real para operadores de contact center.

## 🚀 Funcionalidades Principais

- **Interface Dual Adaptativa:**
    - **Mini View (App):** Janela compacta e nativa focada no essencial (Status, Uptime, Bateria e Relógio).
    - **Full View (Dashboard):** Acessível via navegador, com gráficos de histórico, gestão de logs e configurações.
- **System Tray (Bandeja):** Roda silenciosamente em segundo plano com ícone na bandeja para controle rápido.
- **Autostart Inteligente:** Configuração automática para iniciar com o sistema Linux.
- **Gestão de Identidade:** Nome do operador e cor de identificação persistentes e configuráveis.
- **Monitoramento HID Real:** Captura eventos de botões físicos (Mute, Hook Switch) e detecta conexão/desconexão.
- **Modo de Simulação:** Ativa-se automaticamente na ausência de hardware para facilitar testes de desenvolvimento.
- **Persistência SQLite:** Armazenamento local de configurações, histórico de bateria e logs de hardware.

## 🛠 Stack Técnica

- **Linguagem:** Go (Golang)
- **UI Nativa:** `webview_go` (WebKitGTK)
- **USB HID:** `karalabe/hid`
- **Banco de Dados:** SQLite (`modernc.org/sqlite` - Zero Cgo)
- **Notificações:** `gen2brain/beeep`
- **Frontend:** HTML5, CSS3 moderno e Chart.js (via CDN)

## 📦 Instalação e Desenvolvimento

### Ambiente Nix (Recomendado)
O projeto utiliza Nix Flakes para garantir que todas as dependências nativas (GTK3, WebKitGTK, LibUSB) estejam presentes.
```bash
nix develop
```

### Comandos Úteis
```bash
# Executar em modo desenvolvimento
go run cmd/agent/main.go

# Executar testes de Backend
go test ./internal/...

# Executar testes de Frontend (requer Node.js)
npm install && npm test

# Compilar binário final
go build -o jabra-agent ./cmd/agent/main.go
```

## 🔌 API REST (Porta 18888)

| Método | Endpoint | Descrição |
|--------|----------|-----------|
| `GET` | `/api/telemetry` | Estado atual do dispositivo e operador. |
| `GET` | `/api/history/battery` | Últimos 50 registros de carga da bateria. |
| `GET` | `/api/logs` | Histórico de eventos de hardware (Mute, Botões, etc). |
| `GET` | `/api/config` | Obtém configurações persistentes. |
| `POST` | `/api/config` | Atualiza configurações (Nome, Cor, Autostart). |

## 🧪 Qualidade

O projeto mantém uma cobertura de testes rigorosa:
- **Backend:** Testes unitários para lógica de monitoramento e endpoints de API.
- **Frontend:** Testes de UI via JSDOM para validar a alternância entre Mini e Full view.

---
*Desenvolvido para Marcelo - Aiknow Systems - 2026*