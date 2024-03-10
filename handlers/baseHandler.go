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
	usersInChat []uuid.UUID
}

func NewBaseHandler(pool *db.DB) *BaseHandler {
	return &BaseHandler{
		db:   pool,
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