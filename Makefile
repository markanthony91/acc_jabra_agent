.PHONY: build run clean build-all build-linux build-windows test test-backend test-frontend

BINARY_NAME=acc_jabra_agent

# Build padrão (plataforma atual)
build:
	go build -o $(BINARY_NAME) cmd/agent/main.go

# Executa em modo desenvolvimento
run:
	go run cmd/agent/main.go

# Limpa artefatos de build
clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME).exe $(BINARY_NAME)-*
	rm -rf data/

# Build para Linux AMD64
build-linux:
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME)-linux-amd64 cmd/agent/main.go

# Build para Windows AMD64 (requer MinGW-w64 para CGO)
# -H windowsgui remove console window no Windows
build-windows:
	CGO_ENABLED=1 \
	CC=x86_64-w64-mingw32-gcc \
	GOOS=windows \
	GOARCH=amd64 \
	go build -ldflags="-H windowsgui" -o $(BINARY_NAME).exe cmd/agent/main.go

# Build Windows sem CGO (funcionalidade limitada, sem Jabra SDK)
build-windows-nocgo:
	CGO_ENABLED=0 \
	GOOS=windows \
	GOARCH=amd64 \
	go build -ldflags="-H windowsgui" -o $(BINARY_NAME)-nocgo.exe cmd/agent/main.go

# Build todos os targets
build-all: build-linux build-windows

# Executa testes do backend (Go)
test-backend:
	go test -v ./internal/...

# Executa testes do frontend (Node.js/JSDOM)
test-frontend:
	npm install && npm test

# Executa todos os testes
test: test-backend test-frontend

# Atualiza dependências Go
deps:
	go mod tidy
	go mod download

# Instala dependências do sistema via Nix
setup:
	nix develop

# Copia DLLs necessárias para o diretório de build (Windows)
prepare-windows-dist:
	mkdir -p dist/windows
	cp $(BINARY_NAME).exe dist/windows/
	cp lib/libjabra.dll dist/windows/
	cp -r config dist/windows/
	cp -r public dist/windows/

# Gera arquivo de configuração de exemplo
config-example:
	mkdir -p config
	@echo '{"host":"localhost","port":11967,"token":"","ramal":""}' > config/socket.json
	@echo 'Arquivos de configuração criados em config/'

# Ajuda
help:
	@echo "Targets disponíveis:"
	@echo "  build              - Compila para plataforma atual"
	@echo "  run                - Executa em modo desenvolvimento"
	@echo "  clean              - Remove artefatos de build"
	@echo "  build-linux        - Cross-compile para Linux AMD64"
	@echo "  build-windows      - Cross-compile para Windows AMD64 (requer MinGW)"
	@echo "  build-windows-nocgo- Build Windows sem CGO (sem Jabra SDK)"
	@echo "  build-all          - Compila para todas as plataformas"
	@echo "  test               - Executa todos os testes"
	@echo "  test-backend       - Executa testes Go"
	@echo "  test-frontend      - Executa testes JSDOM"
	@echo "  deps               - Atualiza dependências Go"
	@echo "  setup              - Entra no ambiente Nix"
	@echo "  prepare-windows-dist - Prepara diretório de distribuição Windows"
