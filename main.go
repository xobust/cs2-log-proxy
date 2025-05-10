package main

import (
	"fmt"
	"log"
	"net/http"

	"cs2-log-proxy/domain"
	"cs2-log-proxy/handlers"
	"cs2-log-proxy/storage"
	"cs2-log-proxy/websocket"

	"github.com/gorilla/mux"
)

func main() {
	// Initialize router
	r := mux.NewRouter()

	// Initialize log store
	logStore := storage.NewLogStore("./logs")

	// Initialize WebSocket hub
	hub := websocket.NewHub()

	// WebSocket endpoint
	r.HandleFunc("/ws", websocket.HandleConnections(hub))

	// Domain service
	logService := domain.NewLogService(logStore, hub)

	// API endpoints
	r.HandleFunc("/api/logs", handlers.HandleLogPackage(logService)).Methods("POST")
	r.HandleFunc("/api/logs/{token}", handlers.HandleGetLog(logStore)).Methods("GET")
	r.HandleFunc("/api/listlogs", handlers.HandleListLogs(logService)).Methods("GET")
	r.HandleFunc("/api/config", handlers.HandleConfig).Methods("GET", "POST")

	// Static files for the web UI
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./web")))

	// Start server
	fmt.Println("Starting CS2 Log Manager on :8081")
	if err := http.ListenAndServe(":8081", r); err != nil {
		log.Fatal(err)
	}
}
