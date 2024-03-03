package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	conn   *websocket.Conn
	Token  string `json:"token"`
	chatID string `json:"id"`
}

type Message struct {
	Value string `json:"mess"`
}

func (c *Client) sendMessage(messageType int, key, value string) error {
	message := Message{Value: value}
	messageData, err := json.Marshal(message)
	if err != nil {
		return err
	}
	return c.conn.WriteMessage(messageType, messageData)
}

func (h *BaseHandler) HandleConnections(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	fmt.Println("JWT Token:", token)

	var receivedMessage Client
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&receivedMessage)
	if err != nil {
		log.Println("Error decoding request body:", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
		return
	}

	client := &Client{
		conn:   conn,
		Token:  token,
		chatID: receivedMessage.chatID,
	}

	defer conn.Close()

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		var receivedMessage Message
		err = json.Unmarshal(p, &receivedMessage)
		if err != nil {
			log.Println(err)
			return
		}

		fmt.Printf("Received message - Mess: %s", receivedMessage.Value)

		err = client.sendMessage(messageType, "responseKey", "responseValue")
		if err != nil {
			log.Println(err)
			return
		}
	}
}
