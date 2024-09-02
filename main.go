package main

import (
	"context"
	"log"
	"log/slog"

	"github.com/RobinHood3082/simplebank/api"
	db "github.com/RobinHood3082/simplebank/db/sqlc"
	"github.com/RobinHood3082/simplebank/util"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	conn, err := pgxpool.New(context.Background(), config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	logger := slog.Default()
	store := db.NewStore(conn)
	server := api.NewServer(store, logger, validate)

	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}
