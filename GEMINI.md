# ACC Jabra Agent - GEMINI.md

Este agente é a implementação oficial em Go para telemetria de hardware Jabra, baseada na especificação `acc_jabra_module.md`.

## Visão Geral
- **Linguagem:** Go (Golang)
- **Porta Padrão:** 18888
- **Objetivo:** Capturar eventos de hardware (call, mute, battery) e prover telemetria preditiva.

## Estrutura do Projeto
- `cmd/agent/`: Ponto de entrada.
- `internal/jabra/`: Lógica de integração com hardware (Cgo/HID).
- `internal/api/`: Servidor HTTP REST.
- `internal/models/`: Estruturas de dados compatíveis com ACC Core.

## Comandos Importantes
```bash
# Entrar no ambiente
nix develop

# Executar o agente
go run cmd/agent/main.go

# Build
go build -o agent cmd/agent/main.go
```

## Próximos Passos
- [ ] Implementar integração real via `karalabe/hid` para fallback.
- [ ] Implementar integração Cgo com `Jabra SDK`.
- [ ] Adicionar persistência SQLite para logs de eventos de botões.
