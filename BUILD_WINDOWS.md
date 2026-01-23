# Como compilar e rodar no Windows 11

Este projeto utiliza recursos nativos do Windows (WebView2 e Jabra SDK), por isso a compilação cruzada via Linux é complexa. Recomendamos compilar nativamente no Windows.

## Pré-requisitos

1.  **Go (Golang):** Instale a versão mais recente em [go.dev/dl](https://go.dev/dl/).
2.  **GCC (MinGW-w64):** Necessário para CGO.
    *   Recomendado: Instale via [TDM-GCC](https://jmeubank.github.io/tdm-gcc/) ou via `choco install mingw`.
3.  **Git:** Para clonar e gerenciar o repositório.

## Passo a Passo

1.  **Abra o PowerShell** na pasta do projeto.

2.  **Instale as dependências:**
    ```powershell
    go mod tidy
    ```

3.  **Compile o projeto:**
    ```powershell
    # Compila gerando o executável acc_jabra_agent.exe
    # -H windowsgui remove a janela de console preta (opcional para debug)
    go build -ldflags="-H windowsgui" -o acc_jabra_agent.exe cmd/agent/main.go
    ```

4.  **Verifique os arquivos necessários:**
    Certifique-se de que os seguintes arquivos estão na mesma pasta do executável:
    *   `libjabra.dll` (Copie de `lib/` ou `fastdrive_docs/Lib/`)
    *   Pasta `config/` (com `socket.json`, `keymap.json`)
    *   Pasta `public/` (com `index.html`)

5.  **Execute:**
    Dê dois cliques em `acc_jabra_agent.exe`.

## Solução de Problemas

*   **Erro "EventToken.h not found":** Certifique-se de que o seu GCC está atualizado e suporta headers C++11/17. O TDM-GCC geralmente funciona bem.
*   **Janela não abre:** Verifique se o "WebView2 Runtime" está instalado (padrão no Win11).
*   **Jabra não detectado:** Verifique se o `libjabra.dll` está na mesma pasta e se é a versão correta (64-bit para Go amd64).
