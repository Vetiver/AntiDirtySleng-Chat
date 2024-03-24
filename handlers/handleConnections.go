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
	conn        *websocket.Conn
	UserId      uuid.UUID
	chatID      uuid.UUID
	UsersInChat []uuid.UUID
	messageChan chan *Message
}

type Message struct {
	MessageType int
	ChatId      uuid.UUID
	OwnerID     uuid.UUID
	Value       string `json:"mess"`
}
type ReverseMessage struct {
	Value       string `json:"mess"`
}

func (c *Client) sendMessage(messageType int, value string) error {
	message := ReverseMessage{Value: value}
	messageData, err := json.Marshal(message)
	if err != nil {
		return err
	}
	return c.conn.WriteMessage(messageType, messageData)
}

func (h *BaseHandler) HandleConnections(w http.ResponseWriter, r *http.Request) {
	authToken := r.Header.Get("Authorization")
	token, err := parseToken(authToken)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		log.Println("Invalid token claims")
		http.Error(w, "Invalid token claims", http.StatusUnauthorized)
		return
	}
	userIDStr, ok := claims["id"].(string)
	if !ok || userIDStr == "" {
		log.Println("Missing user ID in token")
		http.Error(w, "Missing user ID in token", http.StatusUnauthorized)
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	params := r.URL.Query()
	chatID, err := uuid.Parse(params.Get("id"))
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid chat ID", http.StatusBadRequest)
		return
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		http.Error(w, "Failed to upgrade connection", http.StatusInternalServerError)
		return
	}
	chatUsers, err := h.db.GetAllUsersInChat(chatID)
	if err != nil {
		log.Println(err)
		conn.Close()
		http.Error(w, "Failed to get chat users", http.StatusInternalServerError)
		return
	}

	client := &Client{
		conn:        conn,
		UserId:      userID,
		chatID:      chatID,
		UsersInChat: chatUsers,
		messageChan: make(chan *Message),
	}
	h.clients[userID] = client
	go func() {
		for {
			messageType, p, err := client.conn.ReadMessage()
			if err != nil {
				log.Println(err)
				return
			}
			var receivedMessage Message
			receivedMessage.MessageType = messageType
			receivedMessage.ChatId = client.chatID
			receivedMessage.OwnerID = client.UserId
			err = json.Unmarshal(p, &receivedMessage)
			if err != nil {
				log.Println(err)
				return
			}
			h.channel <- &receivedMessage
			fmt.Printf("Received message - Mess: %s", receivedMessage.Value)
		}
	}()
	go func() {
		for {
			response := <-client.messageChan
			err = client.sendMessage(response.MessageType, response.Value)
			if err != nil {
				log.Println(err)
				return
			}
			if client.UserId == response.OwnerID {
				err = h.db.InsertMessage(response.ChatId, response.OwnerID, response.Value)
				if err != nil {
					return
				}
			}
		}

	}()
	// defer func() {
	//     h.removeClient(userID)
	// }()
}
