package handlers

import (
	"Chat/db"

	"github.com/google/uuid"
)

type BaseHandler struct {
	db   *db.DB
	usersInChat []uuid.UUID
}

func NewBaseHandler(pool *db.DB) *BaseHandler {
	return &BaseHandler{
		db:   pool,
	}
}