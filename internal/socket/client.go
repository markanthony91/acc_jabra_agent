package socket

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Config contém a configuração do cliente Socket.IO
type Config struct {
	Host  string `json:"host"`
	Port  int    `json:"port"`
	Token string `json:"token"`
	Ramal string `json:"ramal"`

	// Opções de reconexão
	ReconnectInterval time.Duration `json:"reconnect_interval"`
	MaxReconnectTries int           `json:"max_reconnect_tries"`
}

// DefaultConfig retorna configuração padrão
func DefaultConfig() Config {
	return Config{
		Host:              "localhost",
		Port:              11967,
		Token:             "",
		Ramal:             "",
		ReconnectInterval: 5 * time.Second,
		MaxReconnectTries: 10,
	}
}

// Client é o cliente Socket.IO para comunicação com o servidor ACC
type Client struct {
	mu     sync.RWMutex
	config Config
	conn   *websocket.Conn

	connected      bool
	reconnectTries int
	stopChan       chan struct{}

	// Callbacks para eventos recebidos
	onNotificarCarro   func(temCarro bool)
	onLigacaoAtendida  func(ramalQueAtendeu string)
	onLigacaoInterna   func(ramalQueSolicitou string, temCarro bool)
	onConnectionChange func(connected bool)
}

// SocketMessage representa uma mensagem Socket.IO
type SocketMessage struct {
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data"`
}

// ClickPayload é o payload do evento click
type ClickPayload struct {
	Ramal  string `json:"ramal"`
	Token  string `json:"token"`
	Button string `json:"button"`
}

// NotificarCarroPayload é o payload do evento notificar_carro
type NotificarCarroPayload struct {
	TemCarro bool `json:"TemCarro"`
}

// LigacaoAtendidaPayload é o payload do evento ligacao_atendida
type LigacaoAtendidaPayload struct {
	RamalQueAtendeu string `json:"RamalQueAtendeu"`
}

// LigacaoInternaPayload é o payload do evento ligacao_interna
type LigacaoInternaPayload struct {
	RamalQueSolicitou string `json:"RamalQueSolicitou"`
	TemCarro          bool   `json:"TemCarro"`
}

// NewClient cria um novo cliente Socket.IO
func NewClient(config Config) *Client {
	return &Client{
		config:   config,
		stopChan: make(chan struct{}),
	}
}

// Connect estabelece conexão com o servidor
func (c *Client) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected {
		return errors.New("already connected")
	}

	return c.connectInternal()
}

// connectInternal estabelece a conexão (deve ser chamado com lock)
func (c *Client) connectInternal() error {
	// Monta URL do WebSocket
	u := url.URL{
		Scheme: "ws",
		Host:   fmt.Sprintf("%s:%d", c.config.Host, c.config.Port),
		Path:   "/socket.io/",
		RawQuery: url.Values{
			"EIO":       {"4"},
			"transport": {"websocket"},
		}.Encode(),
	}

	// Conecta via WebSocket
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	c.conn = conn
	c.connected = true
	c.reconnectTries = 0

	// Inicia goroutine para ler mensagens
	go c.readLoop()

	// Notifica mudança de conexão
	if c.onConnectionChange != nil {
		go c.onConnectionChange(true)
	}

	log.Printf("[Socket.IO] Conectado a %s:%d", c.config.Host, c.config.Port)
	return nil
}

// Disconnect fecha a conexão
func (c *Client) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil
	}

	close(c.stopChan)
	c.stopChan = make(chan struct{})

	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}

	c.connected = false

	if c.onConnectionChange != nil {
		go c.onConnectionChange(false)
	}

	log.Printf("[Socket.IO] Desconectado")
	return nil
}

// IsConnected retorna o estado da conexão
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// readLoop lê mensagens do servidor
func (c *Client) readLoop() {
	for {
		select {
		case <-c.stopChan:
			return
		default:
		}

		c.mu.RLock()
		conn := c.conn
		c.mu.RUnlock()

		if conn == nil {
			return
		}

		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("[Socket.IO] Erro ao ler mensagem: %v", err)
			c.handleDisconnect()
			return
		}

		c.handleMessage(message)
	}
}

// handleMessage processa uma mensagem recebida
func (c *Client) handleMessage(message []byte) {
	// Socket.IO usa prefixos numéricos para tipos de mensagem
	// 0 = open, 2 = ping, 3 = pong, 4 = message
	if len(message) < 2 {
		return
	}

	// Processa apenas mensagens de evento (prefixo "42")
	if string(message[:2]) != "42" {
		// Responde pings
		if string(message[:1]) == "2" {
			c.sendPong()
		}
		return
	}

	// Remove prefixo "42" e parse JSON
	payload := message[2:]

	// Socket.IO envia eventos como array: ["event_name", data]
	var eventData []json.RawMessage
	if err := json.Unmarshal(payload, &eventData); err != nil {
		log.Printf("[Socket.IO] Erro ao parsear evento: %v", err)
		return
	}

	if len(eventData) < 1 {
		return
	}

	// Extrai nome do evento
	var eventName string
	if err := json.Unmarshal(eventData[0], &eventName); err != nil {
		return
	}

	// Processa eventos conhecidos
	var eventPayload json.RawMessage
	if len(eventData) > 1 {
		eventPayload = eventData[1]
	}

	c.processEvent(eventName, eventPayload)
}

// processEvent processa um evento específico
func (c *Client) processEvent(event string, data json.RawMessage) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	switch event {
	case "notificar_carro":
		if c.onNotificarCarro != nil {
			var payload NotificarCarroPayload
			if err := json.Unmarshal(data, &payload); err == nil {
				go c.onNotificarCarro(payload.TemCarro)
			}
		}

	case "ligacao_atendida":
		if c.onLigacaoAtendida != nil {
			var payload LigacaoAtendidaPayload
			if err := json.Unmarshal(data, &payload); err == nil {
				go c.onLigacaoAtendida(payload.RamalQueAtendeu)
			}
		}

	case "ligacao_interna":
		if c.onLigacaoInterna != nil {
			var payload LigacaoInternaPayload
			if err := json.Unmarshal(data, &payload); err == nil {
				go c.onLigacaoInterna(payload.RamalQueSolicitou, payload.TemCarro)
			}
		}

	default:
		log.Printf("[Socket.IO] Evento desconhecido: %s", event)
	}
}

// handleDisconnect trata desconexão e tenta reconectar
func (c *Client) handleDisconnect() {
	c.mu.Lock()
	c.connected = false
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	c.mu.Unlock()

	if c.onConnectionChange != nil {
		go c.onConnectionChange(false)
	}

	// Tenta reconectar
	go c.reconnect()
}

// reconnect tenta reconectar ao servidor
func (c *Client) reconnect() {
	for {
		c.mu.Lock()
		if c.reconnectTries >= c.config.MaxReconnectTries {
			c.mu.Unlock()
			log.Printf("[Socket.IO] Máximo de tentativas de reconexão atingido")
			return
		}
		c.reconnectTries++
		tries := c.reconnectTries
		c.mu.Unlock()

		log.Printf("[Socket.IO] Tentativa de reconexão %d/%d...", tries, c.config.MaxReconnectTries)

		select {
		case <-c.stopChan:
			return
		case <-time.After(c.config.ReconnectInterval):
		}

		c.mu.Lock()
		err := c.connectInternal()
		c.mu.Unlock()

		if err == nil {
			log.Printf("[Socket.IO] Reconectado com sucesso")
			return
		}

		log.Printf("[Socket.IO] Falha na reconexão: %v", err)
	}
}

// sendPong responde a um ping
func (c *Client) sendPong() {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	if conn != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("3"))
	}
}

// Emit envia um evento ao servidor
func (c *Client) Emit(event string, data interface{}) error {
	c.mu.RLock()
	conn := c.conn
	connected := c.connected
	c.mu.RUnlock()

	if !connected || conn == nil {
		return errors.New("not connected")
	}

	// Socket.IO formato: 42["event_name", data]
	payload := []interface{}{event, data}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	message := append([]byte("42"), jsonData...)
	return conn.WriteMessage(websocket.TextMessage, message)
}

// EmitClick envia evento de click para o servidor
func (c *Client) EmitClick(button string) error {
	payload := ClickPayload{
		Ramal:  c.config.Ramal,
		Token:  c.config.Token,
		Button: button,
	}
	return c.Emit("click", payload)
}

// OnNotificarCarro registra callback para evento notificar_carro
func (c *Client) OnNotificarCarro(handler func(temCarro bool)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onNotificarCarro = handler
}

// OnLigacaoAtendida registra callback para evento ligacao_atendida
func (c *Client) OnLigacaoAtendida(handler func(ramalQueAtendeu string)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onLigacaoAtendida = handler
}

// OnLigacaoInterna registra callback para evento ligacao_interna
func (c *Client) OnLigacaoInterna(handler func(ramalQueSolicitou string, temCarro bool)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onLigacaoInterna = handler
}

// OnConnectionChange registra callback para mudança de estado da conexão
func (c *Client) OnConnectionChange(handler func(connected bool)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onConnectionChange = handler
}

// UpdateConfig atualiza a configuração do cliente
func (c *Client) UpdateConfig(config Config) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.config = config
}

// GetConfig retorna a configuração atual
func (c *Client) GetConfig() Config {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.config
}
