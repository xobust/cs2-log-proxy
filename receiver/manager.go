package receiver

import (
	"context"
	"sync"
	"time"
)

type Receiver struct {
	ID        string
	Type      string
	Config    map[string]interface{}
	Status    string
	LastError error
	LastSeen  time.Time
	mu        sync.Mutex
}

type Manager struct {
	receivers map[string]*Receiver
	mu        sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		receivers: make(map[string]*Receiver),
	}
}

func (m *Manager) AddReceiver(id, typ string, config map[string]interface{}) *Receiver {
	m.mu.Lock()
	defer m.mu.Unlock()

	receiver := &Receiver{
		ID:        id,
		Type:      typ,
		Config:    config,
		Status:    "pending",
		LastError: nil,
	}
	m.receivers[id] = receiver
	return receiver
}

func (m *Manager) GetReceiver(id string) (*Receiver, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	receiver, exists := m.receivers[id]
	return receiver, exists
}

func (m *Manager) UpdateReceiverStatus(id string, status string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if receiver, exists := m.receivers[id]; exists {
		receiver.Status = status
		receiver.LastError = err
		receiver.LastSeen = time.Now()
	}
}

func (m *Manager) ListReceivers() []*Receiver {
	m.mu.RLock()
	defer m.mu.RUnlock()

	receivers := make([]*Receiver, 0, len(m.receivers))
	for _, r := range m.receivers {
		receivers = append(receivers, r)
	}
	return receivers
}

func (m *Manager) ForwardLog(ctx context.Context, log string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, receiver := range m.receivers {
		go func(r *Receiver) {
			// TODO: Implement actual forwarding logic
			r.mu.Lock()
			r.Status = "active"
			r.LastSeen = time.Now()
			r.mu.Unlock()
		}(receiver)
	}
}
