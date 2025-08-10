package pubsub

import (
	"fmt"
	"sync"
)

// Message represents a pub/sub message
type Message struct {
	Topic   string
	Content string
	ID      int64
}

// Listener represents a message handler
type Listener struct {
	ID      int64
	Topic   string
	Handler func(Message)
}

// PubSub represents a simple publish-subscribe system with listeners
type PubSub struct {
	mu        sync.RWMutex
	listeners map[string][]*Listener
	nextID    int64
}

var (
	pubsub *PubSub
	once   sync.Once
	// Global listener registry for C callbacks
	listenerRegistry = make(map[int64]*Listener)
	registryMu       sync.RWMutex
)

// Initialize the pubsub system
func InitPubSub() *PubSub {
	once.Do(func() {
		pubsub = &PubSub{
			listeners: make(map[string][]*Listener),
			nextID:    1,
		}
	})
	return pubsub
}

// Subscribe adds a listener to a topic and returns listener ID
func (ps *PubSub) Subscribe(topic string, handler func(Message)) int64 {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	
	listener := &Listener{
		ID:      ps.nextID,
		Topic:   topic,
		Handler: handler,
	}
	
	ps.nextID++
	ps.listeners[topic] = append(ps.listeners[topic], listener)
	
	// Register globally for C callbacks
	registryMu.Lock()
	listenerRegistry[listener.ID] = listener
	registryMu.Unlock()
	
	return listener.ID
}

// Unsubscribe removes a listener by ID
func (ps *PubSub) Unsubscribe(listenerID int64) bool {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	
	registryMu.Lock()
	listener, exists := listenerRegistry[listenerID]
	if !exists {
		registryMu.Unlock()
		return false
	}
	delete(listenerRegistry, listenerID)
	registryMu.Unlock()
	
	// Remove from topic listeners
	topic := listener.Topic
	listeners := ps.listeners[topic]
	for i, l := range listeners {
		if l.ID == listenerID {
			ps.listeners[topic] = append(listeners[:i], listeners[i+1:]...)
			return true
		}
	}
	
	return false
}

// Publish sends a message to all listeners of a topic
func (ps *PubSub) Publish(topic, content string) {
	ps.mu.RLock()
	listeners := make([]*Listener, len(ps.listeners[topic]))
	copy(listeners, ps.listeners[topic])
	ps.mu.RUnlock()
	
	if len(listeners) == 0 {
		fmt.Printf("No listeners for topic '%s'\n", topic)
		return
	}
	
	message := Message{
		Topic:   topic,
		Content: content,
		ID:      ps.nextID,
	}
	ps.nextID++
	
	// Call all listeners asynchronously
	for _, listener := range listeners {
		go func(l *Listener) {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("Listener %d panicked: %v\n", l.ID, r)
				}
			}()
			l.Handler(message)
		}(listener)
	}
	
	fmt.Printf("Published to '%s' (%d listeners): %s\n", topic, len(listeners), content)
}

// GetListenerCount returns the number of listeners for a topic
func (ps *PubSub) GetListenerCount(topic string) int {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return len(ps.listeners[topic])
}