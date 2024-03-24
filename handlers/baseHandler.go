package handlers

import (
	"Chat/db"
	"fmt"
	"os"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

type BaseHandler struct {
	db   *db.DB
	channel chan *Message
	clients map[uuid.UUID]*Client
}

func NewBaseHandler(pool *db.DB) *BaseHandler {
	h := &BaseHandler{
		db:   pool,
		channel: make(chan *Message),
		clients: make(map[uuid.UUID]*Client),
	}
	go h.updateChans()
	return h
}

func (с *BaseHandler)updateChans() {
	for {
		mess := <- с.channel
		for _, client := range с.clients {
			if client.chatID == mess.ChatId {
                client.messageChan <- mess
            }
		}
	}
}

func (h *BaseHandler) removeClient(userID uuid.UUID) {
    delete(h.clients, userID)
}

func parseToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("error parsing token: %v", err)
	}

	return token, nil
}