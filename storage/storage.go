package storage

import "context"

type LogStorage interface {
	// SaveLog saves a log entry
	SaveLog(ctx context.Context, log string) error

	// GetLogs retrieves logs based on filters
	GetLogs(ctx context.Context, filters map[string]interface{}) ([]string, error)

	// StreamLogs creates a channel for real-time log streaming
	StreamLogs(ctx context.Context) (<-chan string, error)

	// Close closes the storage connection
	Close() error
}

// Config represents the storage configuration
type Config struct {
	Type        string `json:"type"`
	Path        string `json:"path"`
	MaxFileSize int64  `json:"maxFileSize"`
	MaxFiles    int    `json:"maxFiles"`
}
