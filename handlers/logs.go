package handlers

import (
	"cs2-log-proxy/receiver"
	"cs2-log-proxy/storage"
	"cs2-log-proxy/websocket"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// HandleLogPackage handles incoming CS2 log packages
func HandleLogPackage(logStore *storage.LogStore, wsHub *websocket.Hub) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		headers := receiver.CS2ServerHeaders{}

		// Helper to parse int headers
		parseInt := func(key string) int {
			val := r.Header.Get(key)
			if val == "" {
				return 0
			}
			i, _ := strconv.Atoi(val)
			return i
		}

		headers.GameMap = r.Header.Get("X-Game-Map")
		headers.GameScoreCT = parseInt("X-Game-Scorect")
		headers.GameScoreT = parseInt("X-Game-Scoret")
		headers.GameState = r.Header.Get("X-Game-State")
		headers.GameTeamCT = r.Header.Get("X-Game-Teamct")
		headers.GameTeamT = r.Header.Get("X-Game-Teamt")
		headers.LogBytesBeginOffset = parseInt("X-Logbytes-Beginoffset")
		headers.LogBytesEndOffset = parseInt("X-Logbytes-Endoffset")
		headers.ServerAddr = r.Header.Get("X-Server-Addr")
		headers.ServerInstanceToken = r.Header.Get("X-Server-Instance-Token")
		headers.SteamID = r.Header.Get("X-Steamid")
		headers.TickEnd = parseInt("X-Tick-End")
		headers.TickStart = parseInt("X-Tick-Start")
		headers.Timestamp = r.Header.Get("X-Timestamp")

		// Read the POST body as the log chunk
		defer r.Body.Close()
		logData := make([]byte, r.ContentLength)
		_, err := r.Body.Read(logData)
		if err != nil && err.Error() != "EOF" {
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

		// Store the chunk and metadata idempotently
		if logStore != nil {
			err := logStore.SaveChunk(token, string(logData), meta, gameMap, steamID, serverAddr)
			if err != nil {
				http.Error(w, "Failed to save log chunk", http.StatusInternalServerError)
				return
			}
		}
		wsHub.BroadcastEvent("log_chunk", token, string(logData))

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
