package ws_worker

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
)

//type Clients struct {
//	Client1 *Client
//	Client2 *Client
//}

type Client struct {
	Conn     *websocket.Conn
	Message  chan *Message
	ID       string `json:"id"`
	RoomID   string `json:"room_id"`
	Username string `json:"username"`
}

type Message struct {
	Content  string `json:"content"`
	RoomID   string `json:"room_id"`
	Username string `json:"username"`
}

// write message to the client connection
func (c *Client) writeMessage() {
	defer func() {
		err := c.Conn.Close()
		if err != nil {
			fmt.Println("could not close connection")
			return
		}
	}()

	for {
		message, ok := <-c.Message
		if !ok {
			return
		}
		c.Conn.WriteJSON(message)
	}
}

// read message from the client connection and broadcast to the room
func (c *Client) readMessage(hub *Hub) {
	defer func() {
		hub.Unregister <- c
		err := c.Conn.Close()
		if err != nil {
			fmt.Println("could not close connection after read message")
			return
		}
	}()

	for {
		_, m, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		msg := &Message{
			Content:  string(m),
			RoomID:   c.RoomID,
			Username: c.Username,
		}

		// todo distribute the message via redis instead

		hub.Broadcast <- msg
	}
}
