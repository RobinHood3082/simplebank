package main

import (
	"context"
	"log"
	"log/slog"

	"github.com/RobinHood3082/simplebank/internal/app"
	"github.com/RobinHood3082/simplebank/internal/persistence"
	"github.com/RobinHood3082/simplebank/util"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config:", err)
		return
	}

	conn, err := pgxpool.New(context.Background(), config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
		return
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	err = app.SetupValidation(validate)
	if err != nil {
		log.Fatal("cannot setup validation:", err)
		return
	}

	logger := slog.Default()
	store := persistence.NewStore(conn)
	server := app.NewServer(store, logger, validate)

	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}