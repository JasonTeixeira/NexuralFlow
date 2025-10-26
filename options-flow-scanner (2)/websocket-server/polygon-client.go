// ================================================
// POLYGON WEBSOCKET CLIENT
// ================================================
// Connects to Polygon.io WebSocket API
// Streams real-time market data
// Broadcasts to all connected clients
// ================================================

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// ================================================
// POLYGON MESSAGE TYPES
// ================================================

// PolygonMessage represents a message from Polygon
type PolygonMessage struct {
	EventType  string          `json:"ev"`
	Symbol     string          `json:"sym"`
	// Trade fields
	Price      float64         `json:"p"`
	Size       int             `json:"s"`
	Exchange   int             `json:"x"`
	Timestamp  int64           `json:"t"`
	Conditions interface{}     `json:"c"` // Can be int or []int
	// Quote fields
	BidPrice   float64         `json:"bp"`
	BidSize    int             `json:"bs"`
	AskPrice   float64         `json:"ap"`
	AskSize    int             `json:"as"`
	// Aggregate fields
	Open       float64         `json:"o"`
	High       float64         `json:"h"`
	Low        float64         `json:"l"`
	Close      float64         `json:"c"`
	Volume     int64           `json:"v"`
	VWAP       float64         `json:"vw"`
	Raw        json.RawMessage `json:"-"`
}

// PolygonStatusMessage represents a status message
type PolygonStatusMessage struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// PolygonAuthMessage represents auth message
type PolygonAuthMessage struct {
	Action string `json:"action"`
	Params string `json:"params"`
}

// PolygonSubscribeMessage represents subscription message
type PolygonSubscribeMessage struct {
	Action string `json:"action"`
	Params string `json:"params"`
}

// ================================================
// POLYGON CLIENT
// ================================================

type PolygonClient struct {
	apiKey         string
	ws             *websocket.Conn
	connected      bool
	authenticated  bool
	subscriptions  map[string]bool
	messageHandler func(PolygonMessage)
	reconnectTimer *time.Timer
	reconnectDelay time.Duration
	maxRetries     int
	retryCount     int
}

// NewPolygonClient creates a new Polygon WebSocket client
func NewPolygonClient(apiKey string, messageHandler func(PolygonMessage)) *PolygonClient {
	return &PolygonClient{
		apiKey:         apiKey,
		subscriptions:  make(map[string]bool),
		messageHandler: messageHandler,
		reconnectDelay: 5 * time.Second,
		maxRetries:     10,
	}
}

// Connect establishes connection to Polygon WebSocket
func (pc *PolygonClient) Connect() error {
	// Polygon WebSocket URLs:
	// Stocks: wss://socket.polygon.io/stocks
	// Options: wss://socket.polygon.io/options
	// Crypto: wss://socket.polygon.io/crypto
	// Forex: wss://socket.polygon.io/forex
	
	url := "wss://socket.polygon.io/stocks"
	
	log.Printf("🔌 Connecting to Polygon WebSocket: %s", url)
	
	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = 10 * time.Second
	
	ws, _, err := dialer.Dial(url, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to Polygon: %w", err)
	}
	
	pc.ws = ws
	pc.connected = true
	pc.retryCount = 0
	
	log.Println("✅ Connected to Polygon WebSocket")
	
	// Start message reader
	go pc.readMessages()
	
	// Authenticate
	if err := pc.authenticate(); err != nil {
		pc.ws.Close()
		return fmt.Errorf("failed to authenticate: %w", err)
	}
	
	return nil
}

// authenticate sends authentication to Polygon
func (pc *PolygonClient) authenticate() error {
	authMsg := PolygonAuthMessage{
		Action: "auth",
		Params: pc.apiKey,
	}
	
	data, err := json.Marshal(authMsg)
	if err != nil {
		return err
	}
	
	log.Println("🔐 Authenticating with Polygon...")
	
	if err := pc.ws.WriteMessage(websocket.TextMessage, data); err != nil {
		return err
	}
	
	// Wait for auth response (handled in readMessages)
	time.Sleep(2 * time.Second)
	
	if !pc.authenticated {
		return fmt.Errorf("authentication failed")
	}
	
	log.Println("✅ Polygon authentication successful")
	
	return nil
}

// Subscribe subscribes to symbols
func (pc *PolygonClient) Subscribe(symbols []string) error {
	if !pc.connected || !pc.authenticated {
		return fmt.Errorf("not connected or authenticated")
	}
	
	// Build subscription params
	// Format: T.SYMBOL,T.SYMBOL2,... (T = trades, Q = quotes, A = aggregates)
	params := make([]string, 0, len(symbols))
	for _, symbol := range symbols {
		// Subscribe to trades (T), quotes (Q), and aggregates (A)
		params = append(params,
			fmt.Sprintf("T.%s", symbol),  // Trades
			fmt.Sprintf("Q.%s", symbol),  // Quotes
			fmt.Sprintf("A.%s", symbol),  // Aggregates (second bars)
		)
		pc.subscriptions[symbol] = true
	}
	
	subMsg := PolygonSubscribeMessage{
		Action: "subscribe",
		Params: strings.Join(params, ","),
	}
	
	data, err := json.Marshal(subMsg)
	if err != nil {
		return err
	}
	
	log.Printf("📥 Subscribing to Polygon: %v", symbols)
	
	if err := pc.ws.WriteMessage(websocket.TextMessage, data); err != nil {
		return err
	}
	
	return nil
}

// Unsubscribe unsubscribes from symbols
func (pc *PolygonClient) Unsubscribe(symbols []string) error {
	if !pc.connected {
		return fmt.Errorf("not connected")
	}
	
	params := make([]string, 0, len(symbols))
	for _, symbol := range symbols {
		params = append(params,
			fmt.Sprintf("T.%s", symbol),
			fmt.Sprintf("Q.%s", symbol),
			fmt.Sprintf("A.%s", symbol),
		)
		delete(pc.subscriptions, symbol)
	}
	
	unsubMsg := PolygonSubscribeMessage{
		Action: "unsubscribe",
		Params: strings.Join(params, ","),
	}
	
	data, err := json.Marshal(unsubMsg)
	if err != nil {
		return err
	}
	
	log.Printf("📤 Unsubscribing from Polygon: %v", symbols)
	
	if err := pc.ws.WriteMessage(websocket.TextMessage, data); err != nil {
		return err
	}
	
	return nil
}

// readMessages reads messages from Polygon WebSocket
func (pc *PolygonClient) readMessages() {
	defer func() {
		pc.connected = false
		pc.authenticated = false
		if pc.ws != nil {
			pc.ws.Close()
		}
		pc.scheduleReconnect()
	}()
	
	for {
		_, message, err := pc.ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("❌ Polygon WebSocket error: %v", err)
			}
			break
		}
		
		pc.handleMessage(message)
	}
}

// handleMessage processes incoming Polygon messages
func (pc *PolygonClient) handleMessage(data []byte) {
	// Polygon sends arrays of messages
	var messages []json.RawMessage
	if err := json.Unmarshal(data, &messages); err != nil {
		log.Printf("❌ Failed to parse Polygon message: %v", err)
		return
	}
	
	for _, rawMsg := range messages {
		// Check message type
		var msgType struct {
			EventType string `json:"ev"`
			Status    string `json:"status"`
			Message   string `json:"message"`
		}
		
		if err := json.Unmarshal(rawMsg, &msgType); err != nil {
			continue
		}
		
		// Handle status messages
		if msgType.Status != "" {
			pc.handleStatusMessage(msgType.Status, msgType.Message)
			continue
		}
		
		// Handle data messages
		if msgType.EventType != "" {
			pc.handleDataMessage(msgType.EventType, rawMsg)
		}
	}
}

// handleStatusMessage handles Polygon status messages
func (pc *PolygonClient) handleStatusMessage(status, message string) {
	switch status {
	case "auth_success":
		pc.authenticated = true
		log.Println("✅ Polygon: Authentication successful")
		
	case "auth_failed":
		pc.authenticated = false
		log.Printf("❌ Polygon: Authentication failed - %s", message)
		
	case "success":
		log.Printf("✅ Polygon: %s", message)
		
	case "error":
		log.Printf("❌ Polygon: %s", message)
		
	default:
		log.Printf("📨 Polygon status: %s - %s", status, message)
	}
}

// handleDataMessage handles Polygon data messages
func (pc *PolygonClient) handleDataMessage(eventType string, rawMsg json.RawMessage) {
	var msg PolygonMessage
	if err := json.Unmarshal(rawMsg, &msg); err != nil {
		log.Printf("❌ Failed to parse data message: %v", err)
		return
	}
	
	msg.Raw = rawMsg
	
	// Call message handler
	if pc.messageHandler != nil {
		pc.messageHandler(msg)
	}
}

// scheduleReconnect schedules a reconnection attempt
func (pc *PolygonClient) scheduleReconnect() {
	if pc.retryCount >= pc.maxRetries {
		log.Printf("❌ Polygon: Max reconnection attempts reached (%d)", pc.maxRetries)
		return
	}
	
	pc.retryCount++
	delay := pc.reconnectDelay * time.Duration(pc.retryCount)
	
	log.Printf("🔄 Polygon: Reconnecting in %v (attempt %d/%d)", delay, pc.retryCount, pc.maxRetries)
	
	pc.reconnectTimer = time.AfterFunc(delay, func() {
		if err := pc.Connect(); err != nil {
			log.Printf("❌ Polygon reconnection failed: %v", err)
			return
		}
		
		// Resubscribe to all symbols
		symbols := make([]string, 0, len(pc.subscriptions))
		for symbol := range pc.subscriptions {
			symbols = append(symbols, symbol)
		}
		
		if len(symbols) > 0 {
			if err := pc.Subscribe(symbols); err != nil {
				log.Printf("❌ Failed to resubscribe: %v", err)
			}
		}
	})
}

// Disconnect closes the connection
func (pc *PolygonClient) Disconnect() {
	if pc.reconnectTimer != nil {
		pc.reconnectTimer.Stop()
	}
	
	if pc.ws != nil {
		pc.ws.Close()
	}
	
	pc.connected = false
	pc.authenticated = false
	
	log.Println("👋 Disconnected from Polygon")
}

// IsConnected returns connection status
func (pc *PolygonClient) IsConnected() bool {
	return pc.connected && pc.authenticated
}

// GetSubscriptions returns active subscriptions
func (pc *PolygonClient) GetSubscriptions() []string {
	symbols := make([]string, 0, len(pc.subscriptions))
	for symbol := range pc.subscriptions {
		symbols = append(symbols, symbol)
	}
	return symbols
}

// ================================================
// HELPER FUNCTIONS
// ================================================

// TransformPolygonMessage transforms Polygon message to our format
func TransformPolygonMessage(pm PolygonMessage) Message {
	// Determine channel based on event type
	channel := "market-data"
	switch pm.EventType {
	case "T": // Trade
		channel = "trades"
	case "Q": // Quote
		channel = "quotes"
	case "A": // Aggregate (second bar)
		channel = "aggregates"
	case "AM": // Minute aggregate
		channel = "aggregates"
	}
	
	return Message{
		Type:      "market-data",
		Channel:   channel,
		Data:      pm,
		Timestamp: time.Now().UnixMilli(),
		Symbols:   []string{pm.Symbol},
		Metadata: map[string]interface{}{
			"source":     "polygon",
			"event_type": pm.EventType,
			"exchange":   pm.Exchange,
		},
	}
}

// FormatPolygonEventType returns human-readable event type
func FormatPolygonEventType(eventType string) string {
	switch eventType {
	case "T":
		return "Trade"
	case "Q":
		return "Quote"
	case "A":
		return "Second Aggregate"
	case "AM":
		return "Minute Aggregate"
	case "AV":
		return "Value Aggregate"
	default:
		return eventType
	}
}
