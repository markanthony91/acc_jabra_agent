# ACC Jabra Telemetry Agent

Agente de telemetria de hardware escrito em **Go** para monitoramento avançado de headsets Jabra. Parte integrante do ecossistema Aiknow Command Center (ACC), focado em fornecer telemetria em tempo real para operadores de contact center.

## 🖥️ Plataformas Suportadas

| Plataforma | Driver | UI | Status |
|------------|--------|-----|--------|
| **Windows 11** | Jabra SDK (libjabra.dll) | WebView2 | ✅ Principal |
| **Linux** | HID Genérico | WebKitGTK | ⚠️ Fallback |

## 🚀 Funcionalidades Principais

### Interface & UX
- **Interface Dual Adaptativa:**
    - **Mini View (App):** Janela compacta e nativa focada no essencial (Status, Uptime, Bateria e Relógio).
    - **Full View (Dashboard):** Dashboard completo com gráficos de histórico, gestão de logs e configurações.
- **System Tray (Bandeja):** Roda silenciosamente em segundo plano com ícone na bandeja para controle rápido.
- **Autostart Inteligente:** Configuração automática para iniciar com o sistema (Windows Registry / Linux XDG).

### Hardware & Comunicação
- **Jabra SDK Nativo (Windows):** Integração completa via CGO com `libjabra.dll` para controle avançado:
    - Leitura de bateria com status de carregamento
    - Controle de Mute, Ringer, Hook State, Busylight
    - Eventos de botões traduzidos (OffHook, Mute, Volume, etc.)
- **HID Genérico (Linux):** Fallback via `karalabe/hid` para detecção básica.
- **Modo de Simulação:** Ativa-se automaticamente na ausência de hardware.

### Integração Backend
- **Socket.IO Client:** Comunicação em tempo real com servidor ACC:
    - Recebe: `notificar_carro`, `ligacao_atendida`, `ligacao_interna`
    - Envia: `click` (eventos de botão)
- **Motor de Regras (keymap.json):** Mapeamento configurável de botões para ações:
    - `api_call` - Chamadas HTTP
    - `socket_emit` - Eventos Socket.IO
    - `exec` - Comandos do sistema
    - `notify` - Notificações do sistema

### Segurança
- **Device Whitelist:** Lista de dispositivos autorizados por número serial
- **Soft-Block:** Força mute em dispositivos não autorizados
- **Persistência SQLite:** Armazenamento local seguro de configurações

## 🛠 Stack Técnica

| Componente | Windows | Linux |
|------------|---------|-------|
| **Linguagem** | Go (Golang) | Go (Golang) |
| **UI Nativa** | WebView2 (Edge) | WebKitGTK |
| **Hardware** | Jabra SDK (CGO) | karalabe/hid |
| **Banco de Dados** | SQLite (Pure Go) | SQLite (Pure Go) |
| **Notificações** | Windows Toast | D-Bus/Notify |
| **Socket** | gorilla/websocket | gorilla/websocket |

## 📦 Instalação

### Windows 11 (Recomendado)

Para compilar nativamente no Windows, você precisa do compilador GCC instalado (TDM-GCC é recomendado para Go).

1. **Instale o Compilador (TDM-GCC):**
   *   **Download Direto:** [jmeubank.github.io/tdm-gcc/download](https://jmeubank.github.io/tdm-gcc/download/) (Escolha `tdm64-gcc-...exe`)
   *   **Chocolatey:** `choco install mingw`
   *   **Scoop:** `scoop install gcc`
   *   *Dica:* Marque "Add to PATH" durante a instalação.

2. **Baixe o release** ou compile:
   ```powershell
   # Compile nativamente (PowerShell)
   go build -ldflags="-H windowsgui" -o acc_jabra_agent.exe cmd/agent/main.go
   ```

3. **Estrutura necessária:**
   ```
   acc_jabra_agent/
   ├── acc_jabra_agent.exe
   ├── libjabra.dll          # SDK Jabra
   ├── config/
   │   ├── keymap.json       # Mapeamento de botões
   │   ├── socket.json       # Config Socket.IO
   │   └── allowed_devices.json
   └── public/
       └── index.html
   ```

4. **Execute:** Duplo-clique em `acc_jabra_agent.exe`

### Linux (Desenvolvimento)

```bash
# Ambiente Nix (recomendado)
nix develop

# Ou instale manualmente: GTK3, WebKitGTK 4.0, libusb

# Compilar
make build

# Executar
./acc_jabra_agent
```

## ⚙️ Configuração

### config/socket.json
```json
{
  "host": "localhost",
  "port": 11967,
  "token": "SEU_TOKEN",
  "ramal": "12"
}
```

### config/keymap.json
```json
{
  "OffHook": { "action": "socket_emit", "event": "click" },
  "Mute": { "action": "notify", "message": "Mute ativado" },
  "VolumeUp": { "action": "exec", "cmd": "nircmd.exe changesysvolume 5000" }
}
```

### config/allowed_devices.json
```json
{
  "enabled": false,
  "block_mode": "soft",
  "allowed_serials": ["ABC123", "DEF456"]
}
```

## 🔌 API REST (Porta 18888)

| Método | Endpoint | Descrição |
|--------|----------|-----------|
| `GET` | `/api/telemetry` | Estado atual do dispositivo e operador |
| `GET` | `/api/history/battery` | Últimos 50 registros de carga da bateria |
| `GET` | `/api/logs` | Histórico de eventos de hardware |
| `GET` | `/api/config` | Obtém configurações persistentes |
| `POST` | `/api/config` | Atualiza configurações |
| `GET` | `/api/health` | Health check |

## 🔧 Desenvolvimento

### Comandos Make

```bash
make build              # Compila para plataforma atual
make build-windows      # Cross-compile para Windows (requer MinGW)
make build-windows-nocgo # Windows sem Jabra SDK
make build-linux        # Cross-compile para Linux
make test               # Executa todos os testes
make test-backend       # Testes Go
make test-frontend      # Testes JSDOM
make prepare-windows-dist # Prepara distribuição Windows
```

### Estrutura do Projeto

```
acc_jabra_agent/
├── cmd/agent/main.go           # Entry point
├── internal/
│   ├── jabra/
│   │   ├── driver.go           # Interface Driver
│   │   ├── sdk_driver_windows.go  # Jabra SDK (Windows)
│   │   └── hid_driver.go       # HID genérico (Linux)
│   ├── autostart/              # Autostart cross-platform
│   ├── socket/                 # Cliente Socket.IO
│   ├── actions/                # Motor de regras
│   ├── security/               # Device whitelist
│   ├── api/                    # REST API
│   └── db/                     # SQLite persistence
├── lib/
│   ├── JabraSDK.h             # Headers CGO
│   └── libjabra.dll           # SDK Jabra
├── config/                     # Arquivos de configuração
└── public/                     # Frontend HTML/CSS/JS
```

## 🧪 Testes

```bash
# Backend (Go)
go test ./internal/...

# Frontend (Node.js/JSDOM)
npm install && npm test
```

## 📋 Requisitos

### Windows 11
- WebView2 Runtime (incluso no Windows 11)
- `libjabra.dll` junto ao executável
- Headset Jabra compatível

### Linux
- GTK3, WebKitGTK 4.x, libusb
- Ou use `nix develop` para ambiente completo

---
*Desenvolvido para Marcelo - Aiknow Systems - 2026*
