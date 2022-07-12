package websocket

import (
	"bytes"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	// handler function
	handler ClientHandler

	// on websocket close
	closeHandler CloseHandler
}

func (c *Client) Hub() *Hub {
	return c.hub
}

func (c *Client) Broadcast(message []byte) {
	c.hub.broadcast <- message
}

func (c *Client) MultiCast(message []byte) {
	c.hub.multicast <- &multicastMessage{
		c:       c,
		message: message,
	}
}

func (c *Client) SetHub(h *Hub) {
	if c.hub != nil {
		c.hub.unregister <- c
	}
	c.hub = h
	if h != nil {
		h.register <- c
	}
}

func (c *Client) Send(message []byte) {
	c.send <- message
}

func (c *Client) SetCloseHandler(ch CloseHandler) {
	c.closeHandler = ch
}

type ClientHandler func(c *Client, message []byte)

type CloseHandler func(c *Client)

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		if c.hub != nil {
			c.hub.unregister <- c
		}
		c.conn.Close()

		// Call close function when connection closed
		c.closeHandler(c)
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		c.handler(c, message)
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func noCloseHandler(c *Client) {
	return
}

// Default client behavior
var forwardClientHandler ClientHandler = func(c *Client, message []byte) {
	message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
	c.hub.broadcast <- message
}

// serveWs handles websocket requests from the peer.
func ServeWsWithHandler(hub *Hub, w http.ResponseWriter, r *http.Request, h ClientHandler) *Client {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return nil
	}
	client := &Client{
		hub:          hub,
		conn:         conn,
		send:         make(chan []byte, 256),
		handler:      h,
		closeHandler: noCloseHandler,
	}
	if client.hub != nil {
		client.hub.register <- client
	}

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()

	return client
}

// Default functionality preserved
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	ServeWsWithHandler(hub, w, r, forwardClientHandler)
}
