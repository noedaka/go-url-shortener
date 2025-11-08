package audit

import "github.com/noedaka/go-url-shortener/internal/model"

type Observer interface {
	Notify(event model.AuditEvent) error
	Close() error
}

type Subject interface {
	RegisterObserver(observer Observer)
	RemoveObserver(observer Observer)
	NotifyObservers(event model.AuditEvent)
}
