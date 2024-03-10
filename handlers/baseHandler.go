package handlers

import (
	"Chat/db"
	"os"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

type BaseHandler struct {
	db   *db.DB
	UsersInChat []uuid.UUID
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
			client.messageChan <- mess
		}
	}
}



func parseToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	return token, nil
}