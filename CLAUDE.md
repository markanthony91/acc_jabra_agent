# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

ACC Jabra Telemetry Agent is a cross-platform desktop application written in Go that monitors Jabra headsets in real-time. It provides telemetry for contact center operators as part of the Aiknow Command Center (ACC) ecosystem. The codebase is documented in Portuguese.

**Supported Platforms:**
- **Linux** - Uses HID driver via `karalabe/hid`, WebKitGTK for UI
- **Windows 11** - Uses Jabra SDK (`libjabra.dll`) via CGO, WebView2 for UI

## Build and Development Commands

```bash
# Enter Nix environment (Linux - required for GTK/WebKit dependencies)
nix develop

# Run in development mode
go run cmd/agent/main.go

# Run backend tests
go test ./internal/...

# Run frontend tests (JSDOM-based)
npm install && npm test

# Build for current platform
make build

# Build for Linux
make build-linux

# Build for Windows (requires MinGW-w64 for CGO)
make build-windows

# Build for Windows without CGO (limited functionality)
make build-windows-nocgo

# Prepare Windows distribution
make prepare-windows-dist
```

## Architecture

### Multi-Process Design
```
Main Thread (Systray via getlantern/systray)
├── HTTP API Server (goroutine) → Port 18888
├── Socket.IO Client (goroutine) → Backend communication
├── Webview Windows (on-demand)
│   ├── Linux: webview_go + WebKitGTK
│   └── Windows: webview_go + WebView2
└── Hardware Driver (goroutine)
    ├── Linux: HID Scanner (karalabe/hid)
    └── Windows: Jabra SDK (libjabra.dll via CGO)
```

### Key Directories
- `cmd/agent/main.go` - Entry point: systray, webview orchestration
- `internal/api/` - REST API server (port 18888)
- `internal/db/` - SQLite persistence layer
- `internal/jabra/` - Hardware drivers
  - `driver.go` - Abstract Driver interface
  - `sdk_driver_windows.go` - Windows SDK via CGO (build tag: windows)
  - `hid_driver.go` - Linux HID driver (build tag: linux)
  - `monitor.go` - Device monitoring logic
- `internal/autostart/` - Platform-specific autostart
  - `autostart_windows.go` - Windows Registry
  - `autostart_linux.go` - XDG .desktop files
- `internal/socket/` - Socket.IO client for backend communication
- `internal/actions/` - Button action executor (keymap.json)
- `internal/security/` - Device whitelist
- `lib/` - Native libraries
  - `JabraSDK.h` - C headers for CGO
  - `libjabra.dll` - Jabra SDK (Windows)
- `config/` - Configuration files
  - `keymap.json` - Button action mappings
  - `allowed_devices.json` - Device whitelist
- `public/` - Single-page HTML5/CSS3/JS frontend

### Dual-View UI Pattern
The frontend (`public/index.html`) renders two modes based on URL parameter `?view=mini` or `?view=full`:
- **Mini View** (340x380px): Compact native window for battery, status, uptime
- **Full View** (800x600px): Dashboard with Chart.js history, config panel, logs

### Data Flow
```
Hardware Device
    │
    ├── Windows: Jabra SDK (libjabra.dll)
    └── Linux: USB HID (karalabe/hid)
           ↓
    jabra.Driver interface
           ↓
    jabra.Monitor → TelemetryPayload
           ↓
    ├── SQLite Store (persistence)
    ├── REST API endpoints
    ├── Socket.IO (backend events)
    └── Actions Executor (keymap.json)
```

### Simulation Mode
Automatically activates when no Jabra hardware is detected, enabling UI development without physical devices.

## REST API Endpoints

| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/api/telemetry` | Current device state |
| GET | `/api/history/battery` | Last 50 battery readings |
| GET | `/api/logs` | Hardware event history |
| GET | `/api/config` | Persistent settings |
| POST | `/api/config` | Update settings |
| GET | `/api/health` | Health check |

## Socket.IO Events

**Received from server:**
- `notificar_carro` - Drive-thru car notification
- `ligacao_atendida` - Call answered notification
- `ligacao_interna` - Internal call request

**Sent to server:**
- `click` - Button press event with ramal/token

## Configuration Files

### config/keymap.json
Maps Jabra button IDs to actions:
```json
{
  "OffHook": { "action": "socket_emit", "event": "click" },
  "Mute": { "action": "notify", "message": "Mute ativado" }
}
```

Action types: `api_call`, `exec`, `socket_emit`, `notify`, `play_sound`, `none`

### config/allowed_devices.json
Device whitelist for security:
```json
{
  "enabled": false,
  "block_mode": "soft",
  "allowed_serials": ["ABC123"]
}
```

## Key Dependencies

- `webview/webview_go` - Native webview (GTK on Linux, WebView2 on Windows)
- `karalabe/hid` - USB HID device communication (Linux)
- `libjabra.dll` - Jabra SDK (Windows, via CGO)
- `modernc.org/sqlite` - Pure Go SQLite
- `gen2brain/beeep` - System notifications
- `getlantern/systray` - System tray integration
- `gorilla/websocket` - Socket.IO client transport
- `golang.org/x/sys` - Windows Registry access

## System Requirements

**Linux:**
- GTK3, WebKitGTK 4.1, libusb1, pkg-config
- Use `nix develop` for automated setup

**Windows 11:**
- WebView2 Runtime (included in Windows 11)
- `libjabra.dll` in PATH or alongside executable
- MinGW-w64 for building (CGO cross-compilation)

## Jabra SDK (Windows)

The Windows build uses the official Jabra SDK via CGO. Key functions:
- `Jabra_Initialize(appID)` - Initialize with App ID
- `Jabra_GetBatteryStatus()` - Battery level and charging state
- `Jabra_SetMute/SetRinger/SetHookState` - Device control
- Callbacks for button events and device attach/detach

App ID: `88b7-5cbde35c-e588-49b3-a6d5-f54278270e28`
