package audit

import (
    "sync"

    "github.com/noedaka/go-url-shortener/internal/model"
)

type AuditManager struct {
    observers []Observer
    mu        sync.RWMutex
}

func NewAuditManager() *AuditManager {
    return &AuditManager{
        observers: make([]Observer, 0),
    }
}

func (m *AuditManager) RegisterObserver(observer Observer) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.observers = append(m.observers, observer)
}

func (m *AuditManager) RemoveObserver(observer Observer) {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    for i, obs := range m.observers {
        if obs == observer {
            m.observers = append(m.observers[:i], m.observers[i+1:]...)
            break
        }
    }
}

func (m *AuditManager) NotifyObservers(event model.AuditEvent) {
    m.mu.RLock()
    observers := make([]Observer, len(m.observers))
    copy(observers, m.observers)
    m.mu.RUnlock()

    for _, observer := range observers {
        go func(obs Observer) {
            _ = obs.Notify(event)
        }(observer)
    }
}

func (m *AuditManager) Close() {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    for _, observer := range m.observers {
        _ = observer.Close()
    }
    m.observers = nil
}