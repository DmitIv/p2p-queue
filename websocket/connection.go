package websocket

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
	maxBufferSize  = 256
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

type Client struct {
	conn *websocket.Conn

	InboundChan  chan []byte
	OutboundChan chan []byte
}

func NewClient(ctx context.Context, conn *websocket.Conn) (client *Client, err error) {
	client = new(Client)
	client.conn = conn
	client.InboundChan = make(chan []byte, maxBufferSize)
	client.OutboundChan = make(chan []byte, maxBufferSize)

	return client, nil
}

// should be the only one writer per connection
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.InboundChan:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(c.InboundChan)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.InboundChan)
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

func (c *Client) readPump() {
	defer func() {
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		messateType, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("abnormal socket closure: %v\n", err)
			}
			break
		}

		if messateType != websocket.TextMessage {
			log.Println("receive binary data")
		}

		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		c.OutboundChan <- message
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("failed on ws serving: %v\n", err)
		return
	}

	c, err := NewClient(r.Context(), conn)
	if err != nil {
		log.Printf("failed on ws serving: %v\n", err)
		return
	}

	go c.readPump()
	go c.writePump()
}
