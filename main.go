package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"cs2-log-manager/websocket"
	"cs2-log-manager/handlers"
)

func main() {
	// Initialize router
	r := mux.NewRouter()

	// WebSocket endpoint
	r.HandleFunc("/ws", websocket.HandleConnections)

	// API endpoints
	r.HandleFunc("/api/logs", handlers.HandleLogPackage).Methods("POST")
	r.HandleFunc("/api/config", handlers.HandleConfig).Methods("GET", "POST")
	r.HandleFunc("/api/streams", handlers.HandleLogStream).Methods("GET")

	// Static files for the web UI
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./web")))

	// Start message handling goroutine
	go websocket.HandleMessages()

	// Start server
	fmt.Println("Starting CS2 Log Manager on :8081")
	if err := http.ListenAndServe(":8081", r); err != nil {
		log.Fatal(err)
	}
}


