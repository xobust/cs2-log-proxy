package storage

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
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

// ListTokens returns all log tokens by scanning for *_meta.json files
func (ls *LogStore) ListServers() ([]string, error) {
	dirEntries, err := os.ReadDir(ls.Dir)
	if err != nil {
		return nil, err
	}
	tokens := []string{}
	for _, entry := range dirEntries {
		name := entry.Name()
		if len(name) > 7 && name[:7] == "server_" {
			token := name[7 : len(name)-5]
			tokens = append(tokens, token)
		}
	}
	return tokens, nil
}

// LoadServerMeta loads ServerMeta for a given ServerInstanceToken
func (ls *LogStore) LoadServerMeta(token string) (*ServerMeta, error) {
	metaPath := filepath.Join(ls.Dir, "server_"+token+".json")
	var meta ServerMeta
	if f, err := os.Open(metaPath); err == nil {
		defer f.Close()
		if err := json.NewDecoder(f).Decode(&meta); err != nil {
			return nil, err
		}
		return &meta, nil
	} else if os.IsNotExist(err) {
		return &ServerMeta{ServerInstanceToken: token, Logs: []LogMeta{}}, nil
	} else {
		return nil, err
	}
}

// SaveServerMeta writes ServerMeta to disk
func (ls *LogStore) SaveServerMeta(token string, meta *ServerMeta) error {
	metaPath := filepath.Join(ls.Dir, "server_"+token+".json")
	f, err := os.OpenFile(metaPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(meta)
}

// AppendChunk appends chunk data to log file and updates chunk metadata for a given LogID
func (ls *LogStore) AppendChunk(logID string, chunkData string, meta ChunkMeta) error {
	logPath := filepath.Join(ls.Dir, logID+".log")
	metaPath := filepath.Join(ls.Dir, logID+"_chunks.json")

	// Append chunk to log file
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	if _, err := f.WriteString(chunkData); err != nil {
		f.Close()
		return err
	}
	f.Close()

	// Add to metadata and save
	var metas []ChunkMeta
	if mf, err := os.Open(metaPath); err == nil {
		_ = json.NewDecoder(mf).Decode(&metas)
		mf.Close()
	}

	metas = append(metas, meta)
	mf, err := os.OpenFile(metaPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(mf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(metas); err != nil {
		mf.Close()
		return err
	}
	mf.Close()
	return nil
}

// LoadChunkMetas loads all chunk metadata for a given token.
func (ls *LogStore) LoadChunkMetas(token string) ([]ChunkMeta, error) {
	metaPath := filepath.Join(ls.Dir, token+"_chunks.json")
	var metas []ChunkMeta
	if f, err := os.Open(metaPath); err == nil {
		defer f.Close()
		if err := json.NewDecoder(f).Decode(&metas); err != nil {
			return nil, err
		}
	}
	return metas, nil
}

func (ls *LogStore) GetLog(token string) (string, error) {
	logPath := filepath.Join(ls.Dir, token+".log")
	f, err := os.Open(logPath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
