package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type FileStorage struct {
	config    Config
	filePath  string
	file      *os.File
	fileMutex sync.Mutex
	logChan   chan string
	closed    bool
}

func NewFileStorage(config Config) (*FileStorage, error) {
	storage := &FileStorage{
		config:  config,
		logChan: make(chan string, 1000),
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(config.Path, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	// Start log processing goroutine
	go storage.processLogs()

	return storage, nil
}

func (s *FileStorage) SaveLog(ctx context.Context, log string) error {
	select {
	case s.logChan <- log:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *FileStorage) processLogs() {
	for {
		log, ok := <-s.logChan
		if !ok {
			return
		}

		// Create new file if needed
		s.fileMutex.Lock()
		if s.file == nil || s.shouldCreateNewFile() {
			s.closeCurrentFile()
			if err := s.createNewFile(); err != nil {
				fmt.Printf("Error creating new log file: %v\n", err)
				continue
			}
		}
		s.fileMutex.Unlock()

		// Write log
		s.fileMutex.Lock()
		if _, err := s.file.WriteString(log + "\n"); err != nil {
			fmt.Printf("Error writing to log file: %v\n", err)
		}
		s.fileMutex.Unlock()
	}
}

func (s *FileStorage) shouldCreateNewFile() bool {
	if s.file == nil {
		return true
	}

	info, err := s.file.Stat()
	if err != nil {
		return true
	}

	return info.Size() >= s.config.MaxFileSize
}

func (s *FileStorage) createNewFile() error {
	s.filePath = filepath.Join(
		s.config.Path,
		fmt.Sprintf("cs2-log-%s.log", time.Now().Format("2006-01-02_15-04-05")),
	)

	file, err := os.OpenFile(
		s.filePath,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0644,
	)
	if err != nil {
		return fmt.Errorf("failed to create log file: %w", err)
	}

	s.file = file
	return nil
}

func (s *FileStorage) closeCurrentFile() {
	if s.file != nil {
		_ = s.file.Close()
		s.file = nil
	}
}

func (s *FileStorage) GetLogs(ctx context.Context, filters map[string]interface{}) ([]string, error) {
	// TODO: Implement log retrieval with filters
	return nil, nil
}

func (s *FileStorage) StreamLogs(ctx context.Context) (<-chan string, error) {
	out := make(chan string)
	go func() {
		defer close(out)
		for {
			select {
			case log := <-s.logChan:
				out <- log
			case <-ctx.Done():
				return
			}
		}
	}()
	return out, nil
}

func (s *FileStorage) Close() error {
	s.closed = true
	close(s.logChan)
	s.fileMutex.Lock()
	s.closeCurrentFile()
	s.fileMutex.Unlock()
	return nil
}
