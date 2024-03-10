package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	conn   		*websocket.Conn
	UserId  	uuid.UUID
	chatID 		uuid.UUID `json:"id"`
	UsersInChat []uuid.UUID
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
	authToken := r.Header.Get("Authorization")
	fmt.Println("JWT Token:", authToken)
	token, err := parseToken(authToken)
	userID, err := uuid.Parse(token.Claims.(jwt.MapClaims)["id"].(string))
	var receivedMessage Client
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&receivedMessage)
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
	chatUsers, er := h.db.GetAllUsersInChat(receivedMessage.chatID)
	if er != nil {
		log.Println("Error", er)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	client := &Client{
		conn:   conn,
		UserId:  userID,
		chatID: receivedMessage.chatID,
		UsersInChat: chatUsers,
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
