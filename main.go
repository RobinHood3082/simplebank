package main

import (
	"context"
	"log"
	"log/slog"

	"github.com/RobinHood3082/simplebank/api"
	db "github.com/RobinHood3082/simplebank/db/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	dbSource      = "postgresql://backend_stuff:robinrobin@localhost:5432/simple_bank?sslmode=disable"
	serverAddress = "0.0.0.0:8080"
)

func main() {
	conn, err := pgxpool.New(context.Background(), dbSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	logger := slog.Default()
	store := db.NewStore(conn)
	server := api.NewServer(store, logger)

	err = server.Start(serverAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}
