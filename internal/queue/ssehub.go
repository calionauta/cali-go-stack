package queue

import (
	"log/slog"
	"sync"
)

// SSEHub manages Server-Sent Event connections.
// Patterns: register-before-enqueue, replay buffer, backpressure.
type SSEHub struct {
	mu      sync.RWMutex
	clients map[string]chan []byte
	buffer  map[string][][]byte // replay buffer per client
}

func NewSSEHub() *SSEHub {
	return &SSEHub{
		clients: make(map[string]chan []byte),
		buffer:  make(map[string][][]byte),
	}
}

// Register adds a client channel. Must be called BEFORE enqueue.
func (h *SSEHub) Register(clientID string, ch chan []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[clientID] = ch

	// Replay any buffered events
	if buf, ok := h.buffer[clientID]; ok {
		go func() {
			for _, msg := range buf {
				select {
				case ch <- msg:
				default:
					// backpressure: drop if client is slow
				}
			}
		}()
		delete(h.buffer, clientID)
	}
}

// Send pushes data to a specific client. Falls back to buffer if client not ready.
func (h *SSEHub) Send(clientID string, data []byte) {
	h.mu.RLock()
	ch, ok := h.clients[clientID]
	h.mu.RUnlock()

	if ok {
		select {
		case ch <- data:
		default:
			h.bufferEvent(clientID, data)
		}
	} else {
		h.bufferEvent(clientID, data)
	}
}

// Broadcast sends data to all connected clients.
func (h *SSEHub) Broadcast(data []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, ch := range h.clients {
		select {
		case ch <- data:
		default:
			slog.Warn("ssehub: dropping slow client")
		}
	}
}

func (h *SSEHub) bufferEvent(clientID string, data []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	buf := h.buffer[clientID]
	const maxBuffer = 64
	if len(buf) >= maxBuffer {
		buf = buf[1:]
	}
	h.buffer[clientID] = append(buf, data)
}

// Unregister removes a client and its buffer.
func (h *SSEHub) Unregister(clientID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, clientID)
	delete(h.buffer, clientID)
}
