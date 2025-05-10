package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Client represents a websocket client
// Each client can subscribe to multiple event types and tokens
// Subscriptions are managed in the Hub
type Client struct {
	Conn          *websocket.Conn
	Send          chan []byte
	Subscriptions map[string]map[string]bool // eventType -> token -> subscribed
	Mutex         sync.Mutex                 // protects Subscriptions
}

// Hub manages all clients and their subscriptions
type Hub struct {
	Clients       map[*Client]bool
	Subscriptions map[string]map[string]map[*Client]bool // eventType -> token -> set of clients
	Mutex         sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		Clients:       make(map[*Client]bool),
		Subscriptions: make(map[string]map[string]map[*Client]bool),
	}
}

// Message represents a generic message from client
// Supports subscribe/unsubscribe and future event types
type Message struct {
	Type  string `json:"type"`  // "subscribe", "unsubscribe"
	Event string `json:"event"` // e.g. "log_chunk"
	Token string `json:"token"` // log token
}

func (c *Client) writePump() {
	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("Error sending message: %v", err)
				return
			}
		}
	}
}

func (c *Client) readPump(hub *Hub) {
	defer func() {
		hub.RemoveClient(c)
		c.Conn.Close()
		log.Printf("Client disconnected: %s", c.Conn.RemoteAddr().String())
	}()
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}
		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}
		switch msg.Type {
		case "subscribe":
			hub.Subscribe(c, msg.Event, msg.Token)
		case "unsubscribe":
			hub.Unsubscribe(c, msg.Event, msg.Token)
		}
	}
}

// HandleConnections upgrades HTTP to WS and manages client lifecycle
func HandleConnections(hub *Hub) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := Upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("Error upgrading to WebSocket: %v", err)
			return
		}
		client := &Client{
			Conn:          conn,
			Send:          make(chan []byte, 256),
			Subscriptions: make(map[string]map[string]bool),
		}

		hub.AddClient(client)
		log.Printf("Client connected: %s", conn.RemoteAddr().String())

		go client.writePump()
		client.readPump(hub)
	}
}

// AddClient registers a client to the hub
func (h *Hub) AddClient(client *Client) {
	h.Mutex.Lock()
	h.Clients[client] = true
	h.Mutex.Unlock()
}

// RemoveClient unregisters a client and removes all subscriptions
func (h *Hub) RemoveClient(client *Client) {
	h.Mutex.Lock()
	defer h.Mutex.Unlock()
	delete(h.Clients, client)
	for eventType, tokens := range client.Subscriptions {
		for token := range tokens {
			if h.Subscriptions[eventType][token] != nil {
				delete(h.Subscriptions[eventType][token], client)
				if len(h.Subscriptions[eventType][token]) == 0 {
					delete(h.Subscriptions[eventType], token)
				}
			}
		}
	}
}

// Subscribe adds a client's subscription for a given event and token
func (h *Hub) Subscribe(client *Client, eventType, token string) {
	h.Mutex.Lock()
	defer h.Mutex.Unlock()
	if h.Subscriptions[eventType] == nil {
		h.Subscriptions[eventType] = make(map[string]map[*Client]bool)
	}
	if h.Subscriptions[eventType][token] == nil {
		h.Subscriptions[eventType][token] = make(map[*Client]bool)
	}
	h.Subscriptions[eventType][token][client] = true
	client.Mutex.Lock()
	if client.Subscriptions[eventType] == nil {
		client.Subscriptions[eventType] = make(map[string]bool)
	}
	client.Subscriptions[eventType][token] = true
	client.Mutex.Unlock()
}

// Unsubscribe removes a client's subscription for a given event and token
func (h *Hub) Unsubscribe(client *Client, eventType, token string) {
	h.Mutex.Lock()
	defer h.Mutex.Unlock()
	if h.Subscriptions[eventType] != nil && h.Subscriptions[eventType][token] != nil {
		delete(h.Subscriptions[eventType][token], client)
		if len(h.Subscriptions[eventType][token]) == 0 {
			delete(h.Subscriptions[eventType], token)
		}
	}
	client.Mutex.Lock()
	if client.Subscriptions[eventType] != nil {
		delete(client.Subscriptions[eventType], token)
	}
	client.Mutex.Unlock()
}

// BroadcastEvent sends a message to all clients subscribed to a given event and token
func (h *Hub) BroadcastEvent(eventType, token string, payload interface{}) {
	h.Mutex.Lock()
	clients := h.Subscriptions[eventType][token]
	h.Mutex.Unlock()
	if clients == nil {
		return
	}
	msg := map[string]interface{}{
		"type":    eventType,
		"token":   token,
		"payload": payload,
	}
	data, _ := json.Marshal(msg)
	for client := range clients {
		select {
		case client.Send <- data:
		default:
			// Drop message if client is slow
		}
	}
}
