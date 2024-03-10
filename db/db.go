package db

import (
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Client struct {
	Сonn   *websocket.Conn
	Token  string `json:"token"`
	ChatID string `json:"id"`
}

type DB struct {
	pool *pgxpool.Pool
}

func NewDB(pool *pgxpool.Pool) *DB {
	return &DB{
		pool: pool,
	}
}

func DbStart(baseUrl string) *pgxpool.Pool {
	urlExample := baseUrl
	dbpool, err := pgxpool.New(context.Background(), string(urlExample))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v", err)
		os.Exit(1)
	}
	return dbpool
}


func (db DB) chatExists(chatID uuid.UUID) (bool, error) {
	conn, err := db.pool.Acquire(context.Background())
	if err != nil {
		return false, fmt.Errorf("unable to acquire a database connection: %v", err)
	}
	defer conn.Release()

	var exists bool
	err = conn.QueryRow(context.Background(),
		"SELECT EXISTS (SELECT 1 FROM chat WHERE chatid = $1)", chatID).
		Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error checking user existence: %v", err)
	}

	return exists, nil
}

func (db DB) GetAllUsersInChat(chatID uuid.UUID) ([]uuid.UUID, error) {
    exists, err := db.chatExists(chatID)
    if err != nil {
        return nil, err
    }

    if !exists {
        return nil, fmt.Errorf("chat with ID %s does not exist", chatID.String())
    }

    conn, err := db.pool.Acquire(context.Background())
    if err != nil {
        return nil, fmt.Errorf("unable to acquire a database connection: %v", err)
    }
    defer conn.Release()

    rows, err := conn.Query(context.Background(),
        "SELECT \"userid\" FROM user_chat WHERE \"chatid\" = $1", chatID)
    if err != nil {
        return nil, fmt.Errorf("unable to retrieve data from database: %v", err)
    }
    defer rows.Close()

    var data []uuid.UUID
    for rows.Next() {
        var d uuid.UUID
        err = rows.Scan(&d)
        if err != nil {
            return nil, fmt.Errorf("unable to scan row: %v", err)
        }
        data = append(data, d)
    }
    return data, err
}