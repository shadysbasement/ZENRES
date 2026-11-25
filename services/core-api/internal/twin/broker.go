// Package twin fans out live plant / notification events to Server-Sent Event subscribers per org,
// so every open browser tab updates when headcount, batches or the mail bell change.
package twin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Broker struct {
	mu   sync.RWMutex
	subs map[uuid.UUID]map[chan []byte]struct{}
}

func NewBroker() *Broker { return &Broker{subs: map[uuid.UUID]map[chan []byte]struct{}{}} }

func (b *Broker) Publish(org uuid.UUID, event any) {
	data, _ := json.Marshal(event)
	b.mu.RLock()
	defer b.mu.RUnlock()
	for ch := range b.subs[org] {
		select {
		case ch <- data:
		default: // slow consumer: drop rather than block the publisher
		}
	}
}

func (b *Broker) Subscribers(org uuid.UUID) int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.subs[org])
}

// Serve streams events for org until the client disconnects.
func (b *Broker) Serve(w http.ResponseWriter, r *http.Request, org uuid.UUID) {
	fl, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", 500)
		return
	}
	ch := make(chan []byte, 16)
	b.mu.Lock()
	if b.subs[org] == nil {
		b.subs[org] = map[chan []byte]struct{}{}
	}
	b.subs[org][ch] = struct{}{}
	b.mu.Unlock()
	defer func() {
		b.mu.Lock()
		delete(b.subs[org], ch)
		b.mu.Unlock()
	}()
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Accel-Buffering", "no")
	fmt.Fprint(w, ": connected\n\n")
	fl.Flush()
	ping := time.NewTicker(25 * time.Second)
	defer ping.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case <-ping.C:
			fmt.Fprint(w, ": ping\n\n")
			fl.Flush()
		case msg := <-ch:
			fmt.Fprintf(w, "data: %s\n\n", msg)
			fl.Flush()
		}
	}
}
