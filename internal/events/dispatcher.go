package events

import (
	"context"
	"sync"
)

type EventType string

const (
	MenuChangedEvent       EventType = "menu.changed"
	RolePermissionsChanged EventType = "role.permissions.changed"
)

type Event struct {
	Type EventType
	Data interface{}
}

type Listener func(ctx context.Context, event Event)

type Dispatcher interface {
	Dispatch(ctx context.Context, event Event)
	Register(eventType EventType, listener Listener)
}

type dispatcher struct {
	listeners map[EventType][]Listener
	mu        sync.RWMutex
}

var (
	instance Dispatcher
	once     sync.Once
)

func GetDispatcher() Dispatcher {
	once.Do(func() {
		instance = &dispatcher{
			listeners: make(map[EventType][]Listener),
		}
	})
	return instance
}

func (d *dispatcher) Register(eventType EventType, listener Listener) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.listeners[eventType] = append(d.listeners[eventType], listener)
}

func (d *dispatcher) Dispatch(ctx context.Context, event Event) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if listeners, ok := d.listeners[event.Type]; ok {
		for _, listener := range listeners {
			go listener(ctx, event)
		}
	}
}
