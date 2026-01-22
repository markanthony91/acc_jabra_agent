# ACC Jabra Agent - GEMINI.md

Este agente é a implementação oficial em Go para telemetria de hardware Jabra, baseada na especificação `acc_jabra_module.md`.

## Visão Geral
- **Linguagem:** Go (Golang)
- **Porta Padrão:** 18888
- **Objetivo:** Capturar eventos de hardware (call, mute, battery) e prover telemetria preditiva.

## Funcionalidades Implementadas
- [x] Monitoramento de dispositivos HID Jabra via `karalabe/hid`.
- [x] Modo de Simulação (Mock) automático quando hardware não é detectado.
- [x] Dashboard moderno com Webview (Go) e Chart.js para histórico de bateria.
- [x] Persistência de eventos e bateria em SQLite.
- [x] Detecção de Mute e Hook Switch (Call status) via HID.
- [x] System Tray Icon (Bandeja) com menu de contexto (Show/Hide/Exit).
- [x] Início automático (Autostart) no Linux via `.desktop` entry.
- [x] Painel de Configurações para nome do operador, cor e comportamento.
- [x] Visualizador de Logs de Hardware integrado ao Dashboard.
- [x] Persistência de configurações e logs em SQLite.

## Estrutura do Projeto
- `cmd/agent/`: Ponto de entrada (webview + api).
- `internal/jabra/`: Lógica de integração com hardware e simulação.
- `internal/api/`: Servidor HTTP REST para telemetria.
- `internal/db/`: Persistência SQLite (`modernc.org/sqlite`).
- `internal/models/`: Estruturas de dados compatíveis com ACC Core.

## Comandos Importantes
```bash
# Entrar no ambiente
nix develop

# Executar o agente (requer ambiente configurado para WebkitGTK)
go run cmd/agent/main.go

# Build
go build -o agent cmd/agent/main.go
```

## Estatísticas do Projeto
| Métrica | Valor |
|---------|-------|
| Versão | 1.2.0 |
| Backend | Go (Gin-like Stdlib) |
| Frontend | HTML5/Chart.js/JSDOM |
| Testes Unitários | 100% Pass (Back/Front) |
| Horas Estimadas | 16h |
| Portas | 18888 (API/Web) |

## Histórico de Implementação
- [x] Monitoramento de dispositivos HID Jabra via `karalabe/hid`.
- [x] Modo de Simulação (Mock) automático.
- [x] Dashboard moderno com Webview (Go) e Chart.js.
- [x] Persistência de eventos, bateria e settings em SQLite.
- [x] Interface Dual: Mini View (Nativa) e Full View (Web).
- [x] System Tray Icon com menu de controle.
- [x] Autostart no Linux (respeitando configurações).
- [x] API Completa de Configuração e Logs.
