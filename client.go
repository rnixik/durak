package main

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
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

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Client represents a connected user using websockets.
type Client struct {
	lobby *Lobby

	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	isValid bool

	nickname string
	id       uint64
	room     *Room
}

// ClientCommand is a command message from connected client.
type ClientCommand struct {
	Type    string      `json:"type"`
	SubType string      `json:"sub_type"`
	Data    interface{} `json:"data"`
	client  *Client
}

// Nickname returns nickname of the client
func (c *Client) Nickname() string {
	return c.nickname
}

func (c *Client) readLoop() {
	defer func() {
		c.lobby.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("error: %v", err)
			}
			break
		}
		log.Printf("Incoming message: %s", message)

		var clientCommand ClientCommand
		if err := json.Unmarshal(message, &clientCommand); err != nil {
			log.Printf("json unmarshal error: %s", err)
		} else {
			clientCommand.client = c
			c.lobby.onClientCommand(&clientCommand)
		}
	}
}

func (c *Client) writeLoop() {
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
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

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

func (c *Client) sendEvent(event interface{}) {
	json, _ := eventToJSON(event)
	log.Printf("Send json to client %s", c.nickname)
	log.Println(string(json))
	c.sendMessage(json)
}

func (c *Client) sendMessage(message []byte) {
	if c.send == nil {
		return
	}
	c.send <- message
}

func serveWs(lobby *Lobby, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &Client{
		lobby: lobby,
		conn:  conn,
		send:  make(chan []byte, 256),
	}
	client.lobby.register <- client

	go client.writeLoop()
	go client.readLoop()
}
