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
	ChatId 		uuid.UUID
	Value       string `json:"mess"`
}

func (c *Client) sendMessage(messageType int, value string) error {
	messageData, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.conn.WriteMessage(messageType, messageData)
}

func (h *BaseHandler) HandleConnections(w http.ResponseWriter, r *http.Request) {
	authToken := r.Header.Get("Authorization")
	fmt.Println("JWT Token:", authToken)
	token, err := parseToken(authToken)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}
	userID, err := uuid.Parse(token.Claims.(jwt.MapClaims)["id"].(string))
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
			response := <- client.messageChan
				err = client.sendMessage(response.MessageType, response.Value)
				if err != nil {
					log.Println(err)
					return
				}
		}
		
	}()
	defer func() {
        h.removeClient(userID)
    }()
}
