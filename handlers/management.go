package handlers

import (
	"cs2-log-proxy/storage"
	"encoding/json"
	"net/http"
)

// ManagementHandler handles admin APIs (log listing, etc)
// Route: /api/admin/*
func ManagementHandler(logStore *storage.LogStore) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/admin/logs", func(w http.ResponseWriter, r *http.Request) {
		logs, err := logStore.ListLogs()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Failed to list logs"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(logs)
	})
	// Add more admin endpoints here
	return mux
}
