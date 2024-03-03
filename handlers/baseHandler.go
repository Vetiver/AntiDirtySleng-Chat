package handlers

import (
	"Chat/db"
)

type BaseHandler struct {
	db   *db.DB

}

func NewBaseHandler(pool *db.DB) *BaseHandler {
	return &BaseHandler{
		db:   pool,
	}
}