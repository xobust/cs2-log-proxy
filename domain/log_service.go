package domain

import (
	"log"
	"sort"
	"strings"
	"time"

	"cs2-log-proxy/storage"
	"cs2-log-proxy/websocket"
)

type LogService struct {
	Store *storage.LogStore
	Hub   *websocket.Hub
}

// LogSummary holds summary info for listing logs
// LastActivity is the timestamp of the latest chunk (ISO8601 string)
type LogSummary struct {
	LogID        string              `json:"log_id"`
	Token        string              `json:"server_instance_token"`
	LogStartTime string              `json:"log_start_time"`
	LogMetadata  storage.LogMetadata `json:"metadata"`
	LastActivity string              `json:"last_activity"`
}

func NewLogService(store *storage.LogStore, hub *websocket.Hub) *LogService {
	return &LogService{Store: store, Hub: hub}
}

// ProcessLogChunk checks for new logs, chunk overlaps, and triggers events.
func (svc *LogService) ProcessLogChunk(token string, chunkData string, meta storage.ChunkMeta, gameMap, steamID, serverAddr string) (bool, error) {
	// Load or create ServerMeta
	serverMeta, err := svc.Store.LoadServerMeta(token)
	if err != nil {
		return false, err
	}

	var logId string = ""
	var isNewLog bool = false

	for i, log := range serverMeta.Logs {
		if log.LastByteOffset == meta.BeginOffset && TimestampDiff(log.LastActivity, meta.Timestamp) < time.Hour*2 && log.GameMap == gameMap {
			logId = log.LogID
			serverMeta.Logs[i].LastActivity = meta.Timestamp
			serverMeta.Logs[i].LastByteOffset = meta.EndOffset
			break
		}
	}

	if logId == "" {
		if meta.BeginOffset != 0 {
			// Warn create new log from non-zero offset
			log.Printf("Creating new log from non-zero offset: %d", meta.BeginOffset)
		}

		logId = token + "_" + strings.ReplaceAll(meta.Timestamp, "/", "_")

		newLog := storage.LogMeta{
			LogID:          logId,
			LogStartTime:   meta.Timestamp,
			GameMap:        gameMap,
			ServerAddr:     serverAddr,
			LastActivity:   meta.Timestamp,
			LastByteOffset: meta.EndOffset,
		}
		serverMeta.Logs = append(serverMeta.Logs, newLog)
		serverMeta.SteamID = steamID
		if err := svc.Store.SaveServerMeta(token, serverMeta); err != nil {
			return false, err
		}
		isNewLog = true
	}

	metas, err := svc.Store.LoadChunkMetas(logId)
	if err != nil {
		return false, err
	}

	shouldSave := true
	chunkToSave := chunkData
	metaToSave := meta

	// Check if chunk is new or overlapping
	for _, m := range metas {
		if m.BeginOffset == meta.BeginOffset {
			if meta.EndOffset > m.EndOffset {
				log.Printf("Overlapping chunk: %d", meta.BeginOffset)
				// Overlapping, but new chunk extends further: split and save only the new part
				newPart, newMeta := splitChunk(m.EndOffset, meta, chunkData)
				if newPart == "" {
					shouldSave = false
				} else {
					chunkToSave = newPart
					metaToSave = newMeta
				}
			} else {
				// Duplicate or less complete chunk, ignore
				shouldSave = false
			}
			break
		}
	}

	if shouldSave {
		if err := svc.Store.AppendChunk(logId, chunkToSave, metaToSave); err != nil {
			return false, err
		}
		if err := svc.Store.SaveServerMeta(token, serverMeta); err != nil {
			return false, err
		}
		svc.Hub.BroadcastEvent("log_chunk", logId, chunkToSave)
	}

	if isNewLog {
		summary := LogSummary{
			Token:        token,
			LogID:        logId,
			LogStartTime: meta.Timestamp,
			LogMetadata: storage.LogMetadata{
				ServerInstanceToken: token,
				GameMap:             gameMap,
				SteamID:             steamID,
				ServerAddr:          serverAddr,
			},
			LastActivity: meta.Timestamp,
		}
		log.Printf("New log: %v", summary)
		svc.Hub.BroadcastEvent("new_log", "*", summary)
	}

	return isNewLog, nil
}

// ListLogs returns all logs with metadata and last activity, ordered by last activity desc
func (svc *LogService) ListLogs() ([]LogSummary, error) {
	tokens, err := svc.Store.ListServers()
	if err != nil {
		return []LogSummary{}, err
	}
	var result []LogSummary = []LogSummary{}
	for _, serverId := range tokens {
		meta, err := svc.Store.LoadServerMeta(serverId)
		if err != nil {
			continue
		}
		for _, log := range meta.Logs {
			result = append(result, LogSummary{
				Token:        serverId,
				LogID:        log.LogID,
				LogStartTime: log.LogStartTime,
				LogMetadata: storage.LogMetadata{
					ServerInstanceToken: serverId,
					GameMap:             log.GameMap,
					ServerAddr:          log.ServerAddr,
					SteamID:             meta.SteamID,
				},
				LastActivity: log.LastActivity,
			})
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].LastActivity > result[j].LastActivity
	})
	return result, nil
}

// IsNewLog returns true if this is the first chunk for the token
func (svc *LogService) IsNewLog(token string) (bool, error) {
	metas, err := svc.Store.LoadChunkMetas(token)
	if err != nil {
		return false, err
	}
	return len(metas) == 0, nil
}

// splitChunk returns only the new part of the chunk and adjusted meta
func splitChunk(existingEnd int, meta storage.ChunkMeta, chunkData string) (string, storage.ChunkMeta) {
	// Assume chunkData is contiguous log text, and offsets refer to byte positions
	start := existingEnd - meta.BeginOffset
	if start < 0 || start >= len(chunkData) {
		return "", meta // nothing new
	}
	return chunkData[start:], storage.ChunkMeta{
		BeginOffset: existingEnd,
		EndOffset:   meta.EndOffset,
		GameScoreCT: meta.GameScoreCT,
		GameScoreT:  meta.GameScoreT,
		GameState:   meta.GameState,
		GameTeamCT:  meta.GameTeamCT,
		GameTeamT:   meta.GameTeamT,
		TickEnd:     meta.TickEnd,
		TickStart:   meta.TickStart,
		Timestamp:   meta.Timestamp,
	}
}

func TimestampDiff(first, second string) time.Duration {
	layout := "01/02/2006 - 15:04:05.000"
	firstTime, _ := time.Parse(layout, first)
	secondTime, _ := time.Parse(layout, second)
	return secondTime.Sub(firstTime)
}
