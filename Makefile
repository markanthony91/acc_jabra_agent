.PHONY: build run clean build-all

BINARY_NAME=acc_jabra_agent

build:
	go build -o $(BINARY_NAME) cmd/agent/main.go

run:
	go run cmd/agent/main.go

clean:
	rm -f $(BINARY_NAME)
	rm -rf data/

build-linux:
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME)-linux-amd64 cmd/agent/main.go

# Instala dependências do sistema necessárias para compilação (Nix)
setup:
	nix develop
