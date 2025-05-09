package websocket

import (
	"log"
	"net/http"
	"sync"
	"github.com/gorilla/websocket"
)

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	Conn *websocket.Conn
}

var Clients = make(map[*Client]bool)
var Broadcast = make(chan string)
var Mutex = &sync.Mutex{}

func HandleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading to WebSocket: %v", err)
		return
	}

	client := &Client{conn}
	Mutex.Lock()
	Clients[client] = true
	Mutex.Unlock()

	log.Printf("Client connected: %s", conn.RemoteAddr().String())

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			Mutex.Lock()
			delete(Clients, client)
			Mutex.Unlock()
			log.Printf("Client disconnected: %s", conn.RemoteAddr().String())
			break
		}
	}

	conn.Close()
}

func HandleMessages() {
	for {
		msg := <-Broadcast
		Mutex.Lock()
		for client := range Clients {
			err := client.Conn.WriteMessage(websocket.TextMessage, []byte(msg))
			if err != nil {
				log.Printf("Error sending message: %v", err)
				client.Conn.Close()
				delete(Clients, client)
			}
		}
		Mutex.Unlock()
	}
}
