package handlers

import (
	"cs2-log-proxy/domain"
	"cs2-log-proxy/storage"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mozillazg/go-httpheader"
)

// CS2ServerHeaders represents all custom server-specific headers from CS2 log POSTs
// Add more fields as needed if new headers appear in the future.
type CS2ServerHeaders struct {
	GameMap             string `header:"X-Game-Map"`
	GameScoreCT         int    `header:"X-Game-Scorect"`
	GameScoreT          int    `header:"X-Game-Scoret"`
	GameState           string `header:"X-Game-State"`
	GameTeamCT          string `header:"X-Game-Teamct"`
	GameTeamT           string `header:"X-Game-Teamt"`
	LogBytesBeginOffset int    `header:"X-Logbytes-Beginoffset"`
	LogBytesEndOffset   int    `header:"X-Logbytes-Endoffset"`
	ServerAddr          string `header:"X-Server-Addr"`                     // IP address of the server
	ServerInstanceToken string `header:"X-Server-Instance-Token,omitempty"` // Unique ID of this match
	ServerUniqueToken   string `header:"X-Server-Unique-Token"`             // CRC64 hash of logaddress_token_secret
	SteamID             string `header:"X-Steamid"`
	TickEnd             int    `header:"X-Tick-End"`
	TickStart           int    `header:"X-Tick-Start"`
	Timestamp           string `header:"X-Timestamp"` // MM/DD/YYYY - HH:MM:SS.MMM Example: 01/30/2025 - 16:33:56.470
}

// HandleLogPackage handles incoming CS2 log packages
func HandleLogPackage(logService *domain.LogService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		// Parse headers into struct
		headers := CS2ServerHeaders{}
		if err := httpheader.Decode(r.Header, &headers); err != nil {
			http.Error(w, "Failed to parse headers", http.StatusBadRequest)
			return
		}

		if int(r.ContentLength) != headers.LogBytesEndOffset-headers.LogBytesBeginOffset {
			log.Printf("Content length mismatch: %d != %d", r.ContentLength, headers.LogBytesEndOffset-headers.LogBytesBeginOffset)
		}

		// Read the POST body as the log chunk
		defer r.Body.Close()
		logData := make([]byte, r.ContentLength)
		n, err := r.Body.Read(logData)
		if err != nil && err.Error() != "EOF" {
			http.Error(w, "Failed to read log chunk", http.StatusBadRequest)
			return
		}
		if n != int(r.ContentLength) {
			log.Printf("Content length mismatch: %d != %d", n, r.ContentLength)
			log.Printf("This somtimes happens during start, recovers after log gets appended")
			http.Error(w, "Failed to read log chunk", http.StatusBadRequest)
			return
		}

		// Prepare chunk metadata (all header fields except ServerInstanceToken and GameMap)
		meta := storage.ChunkMeta{
			BeginOffset: headers.LogBytesBeginOffset,
			EndOffset:   headers.LogBytesEndOffset,
			GameScoreCT: headers.GameScoreCT,
			GameScoreT:  headers.GameScoreT,
			GameState:   headers.GameState,
			GameTeamCT:  headers.GameTeamCT,
			GameTeamT:   headers.GameTeamT,
			TickEnd:     headers.TickEnd,
			TickStart:   headers.TickStart,
			Timestamp:   headers.Timestamp,
		}
		token := headers.ServerInstanceToken
		if token == "" {
			http.Error(w, "Missing ServerInstanceToken", http.StatusBadRequest)
			return
		}
		gameMap := headers.GameMap
		steamID := headers.SteamID
		serverAddr := headers.ServerAddr

		if logService == nil {
			http.Error(w, "Log service unavailable", http.StatusInternalServerError)
			return
		}
		_, err = logService.ProcessLogChunk(token, string(logData), meta, gameMap, steamID, serverAddr)
		if err != nil {
			log.Printf("Failed to process log chunk: %v", err)
			http.Error(w, "Failed to process log chunk", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func HandleGetLog(logStore *storage.LogStore) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		token := mux.Vars(r)["token"]
		if token == "" {
			http.Error(w, "Missing token", http.StatusBadRequest)
			return
		}
		logs, err := logStore.GetLog(token)
		if err != nil {
			http.Error(w, "Failed to get log", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(logs))
	}
}

func HandleListLogs(logService *domain.LogService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logs, err := logService.ListLogs()
		if err != nil {
			http.Error(w, "Failed to list logs", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(logs)
	}
}
