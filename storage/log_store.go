package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

// LogStore manages log file and chunk metadata for each ServerInstanceToken
type LogStore struct {
	Dir      string
	mutexMap map[string]*sync.Mutex // per-token mutex
	mu       sync.Mutex             // guards mutexMap
}

func NewLogStore(dir string) *LogStore {
	return &LogStore{
		Dir:      dir,
		mutexMap: make(map[string]*sync.Mutex),
	}
}

func (ls *LogStore) getMutex(token string) *sync.Mutex {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	m, ok := ls.mutexMap[token]
	if !ok {
		m = &sync.Mutex{}
		ls.mutexMap[token] = m
	}
	return m
}

// SaveLogMetadata writes metadata for a log (ServerInstanceToken, GameMap)
func (ls *LogStore) SaveLogMetadata(token string, meta LogMetadata) error {
	metaPath := filepath.Join(ls.Dir, token+"_meta.json")
	f, err := os.OpenFile(metaPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(meta)
}

// GetLogMetadata loads metadata for a log
func (ls *LogStore) GetLogMetadata(token string) (*LogMetadata, error) {
	metaPath := filepath.Join(ls.Dir, token+"_meta.json")
	f, err := os.Open(metaPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var meta LogMetadata
	if err := json.NewDecoder(f).Decode(&meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

// LogSummary holds summary info for listing logs
// LastActivity is the timestamp of the latest chunk (ISO8601 string)
type LogSummary struct {
	Token        string      `json:"server_instance_token"`
	LogMetadata  LogMetadata `json:"metadata"`
	LastActivity string      `json:"last_activity"`
}

// ListLogs returns all logs with metadata and last activity, ordered by last activity desc
func (ls *LogStore) ListLogs() ([]LogSummary, error) {
	dirEntries, err := os.ReadDir(ls.Dir)
	if err != nil {
		return []LogSummary{}, err // Always return an array, not nil
	}
	tokenSet := make(map[string]struct{})
	for _, entry := range dirEntries {
		name := entry.Name()
		if len(name) > 10 && name[len(name)-10:] == "_meta.json" {
			token := name[:len(name)-10]
			tokenSet[token] = struct{}{}
		}
	}
	var result []LogSummary = []LogSummary{}
	for token := range tokenSet {
		meta, err := ls.GetLogMetadata(token)
		if err != nil {
			continue
		}
		lastActivity := ""
		chunkPath := filepath.Join(ls.Dir, token+"_chunks.json")
		if f, err := os.Open(chunkPath); err == nil {
			var chunks []ChunkMeta
			if err := json.NewDecoder(f).Decode(&chunks); err == nil && len(chunks) > 0 {
				lastActivity = chunks[len(chunks)-1].Timestamp
			}
			f.Close()
		}
		result = append(result, LogSummary{
			Token:        token,
			LogMetadata:  *meta,
			LastActivity: lastActivity,
		})
	}
	// Sort by lastActivity desc
	sort.Slice(result, func(i, j int) bool {
		return result[i].LastActivity > result[j].LastActivity
	})
	return result, nil
}

// SaveChunk appends chunk data to log file and updates chunk metadata (idempotent)
// If log metadata doesn't exist, create it from chunk (requires GameMap, SteamID, ServerAddr)
func (ls *LogStore) SaveChunk(token string, chunkData string, meta ChunkMeta, gameMap, steamID, serverAddr string) error {
	mutex := ls.getMutex(token)
	mutex.Lock()
	defer mutex.Unlock()

	logPath := filepath.Join(ls.Dir, token+".log")
	metaPath := filepath.Join(ls.Dir, token+"_chunks.json")
	logMetaPath := filepath.Join(ls.Dir, token+"_meta.json")

	// If log metadata doesn't exist, create it
	if _, err := os.Stat(logMetaPath); os.IsNotExist(err) {
		logMeta := LogMetadata{
			ServerInstanceToken: token,
			GameMap:             gameMap,
			SteamID:             steamID,
			ServerAddr:          serverAddr,
		}
		_ = ls.SaveLogMetadata(token, logMeta)
	}

	// Load chunk metadata
	var metas []ChunkMeta
	if f, err := os.Open(metaPath); err == nil {
		_ = json.NewDecoder(f).Decode(&metas)
		f.Close()
	}
	// Check for idempotency (skip if offset exists)
	for _, m := range metas {
		if m.BeginOffset == meta.BeginOffset && m.EndOffset == meta.EndOffset {
			return nil // Already stored
		}
	}
	// Append chunk to log file
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("open log: %w", err)
	}
	if _, err := f.WriteString(chunkData); err != nil {
		f.Close()
		return fmt.Errorf("write log: %w", err)
	}
	f.Close()
	// Add to metadata and save
	metas = append(metas, meta)
	mf, err := os.OpenFile(metaPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("open meta: %w", err)
	}
	enc := json.NewEncoder(mf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(metas); err != nil {
		mf.Close()
		return fmt.Errorf("write meta: %w", err)
	}
	mf.Close()
	return nil
}
