package http

import (
	"fmt"
	"github.com/gorilla/websocket"
	"sync"
)

// wrapper around a websocket
type websocketListener struct {
	sync.Mutex
	Conn   *websocket.Conn
	closed bool
}

func wrapWebSocket(c *websocket.Conn) *websocketListener {
	res := &websocketListener{
		Conn:   c,
		closed: false,
	}

	c.SetCloseHandler(res.onclose)

	go func() {
		for {
			if _, _, err := c.NextReader(); err != nil {
				res.Close()
				break
			}
		}
	}()

	return res
}

func (wl *websocketListener) onclose(code int, text string) error {
	fmt.Println("Websocket closed", code, text)

	wl.Close()

	return nil
}

func (wl *websocketListener) ReceiveMessage(msg string) bool {
	wl.Lock()
	defer wl.Unlock()

	if wl.closed {
		return false
	}

	wl.Conn.WriteJSON(msg)

	return true
}

func (wl *websocketListener) Close() {
	wl.Lock()
	defer wl.Unlock()

	wl.Conn.Close()
	wl.closed = true
}
