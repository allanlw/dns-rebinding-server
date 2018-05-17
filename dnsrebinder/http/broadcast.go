package http

import (
	"fmt"
	"sync"
)

// Return false to stop receiving new messages
type subscriber interface {
	ReceiveMessage(string) bool
}

var listeners = struct {
	sync.Mutex
	subscribers []subscriber
}{}

type BroadcastWriter struct{}

func (b *BroadcastWriter) Write(p []byte) (n int, err error) {
	BroadcastToListeners(fmt.Sprintf("%s", p))

	return len(p), nil
}

func addListener(s subscriber) {
	listeners.Lock()
	defer listeners.Unlock()

	listeners.subscribers = append(listeners.subscribers, s)
}

func BroadcastToListeners(msg string) {
	listeners.Lock()
	defer listeners.Unlock()

	var newSubs []subscriber

	for _, listener := range listeners.subscribers {
		if listener.ReceiveMessage(msg) {
			newSubs = append(newSubs, listener)
		}
	}

	listeners.subscribers = newSubs
}
