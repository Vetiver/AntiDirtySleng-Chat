package main

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
	conn *websocket.Conn
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

func handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
		return
	}
	client := &Client{conn: conn}
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

func main() {
	http.HandleFunc("/ws", handleConnections)

	log.Println("Server started on :8000")
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("Error starting server: ", err)
	}
}
