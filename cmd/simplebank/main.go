package main

import (
	"context"
	"log"
	"log/slog"
	"net"
	"net/http"

	"github.com/RobinHood3082/simplebank/internal/app"
	"github.com/RobinHood3082/simplebank/internal/gapi"
	"github.com/RobinHood3082/simplebank/internal/pb"
	"github.com/RobinHood3082/simplebank/internal/persistence"
	"github.com/RobinHood3082/simplebank/internal/token"
	"github.com/RobinHood3082/simplebank/util"
	"github.com/go-playground/validator/v10"
	"github.com/golang-migrate/migrate/v4"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rakyll/statik/fs"
	"google.golang.org/protobuf/encoding/protojson"

	_ "github.com/RobinHood3082/simplebank/doc/statik"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
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

	// Run db migrations
	err = runDBMigrations(config.MigrationURL, config.DBSource)
	if err != nil {
		log.Fatal("cannot run db migrations:", err)
		return
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	err = app.SetupValidation(validate)
	if err != nil {
		log.Fatal("cannot setup validation:", err)
		return
	}

	var tokenMaker token.Maker
	switch config.TokenType {
	case "jwt":
		tokenMaker, err = token.NewJWTMaker(config.TokenSymmetricKey)
		if err != nil {
			log.Fatal("cannot create token maker:", err)
			return
		}
	case "paseto":
		tokenMaker, err = token.NewPasetoMaker(config.TokenSymmetricKey)
		if err != nil {
			log.Fatal("cannot create token maker:", err)
			return
		}
	default:
		log.Fatal("unknown token type")
		return
	}

	logger := slog.Default()
	store := persistence.NewStore(conn)

	go func() {
		err := runGatewayServer(store, logger, tokenMaker, config)
		if err != nil {
			log.Fatal("failed to run gateway server:", err)
		}
	}()

	err = runGRPCServer(store, logger, tokenMaker, config)
	if err != nil {
		log.Fatal("exiting server application")
		return
	}

}

func runDBMigrations(migrationURL string, dbSource string) error {
	migrations, err := migrate.New(migrationURL, dbSource)
	if err != nil {
		log.Fatal("cannot create new migrate instance:", err)
		return err
	}

	err = migrations.Up()
	if err != nil && err != migrate.ErrNoChange {
		log.Fatal("failed to run up migrations:", err)
		return err
	}

	log.Println("db migrated successfully")
	return nil
}

func runGRPCServer(store persistence.Store, logger *slog.Logger, tokenMaker token.Maker, config util.Config) error {
	server := gapi.NewServer(store, logger, tokenMaker, config)

	err := server.Start(config.GRPCServerAddress)
	if err != nil {
		log.Fatal("cannot start gRPC server:", err)
		return err
	}

	return nil
}

func runGatewayServer(store persistence.Store, logger *slog.Logger, tokenMaker token.Maker, config util.Config) error {
	server := gapi.NewServer(store, logger, tokenMaker, config)

	jsonOption := runtime.WithMarshalerOption(
		runtime.MIMEWildcard,
		&runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames: true,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		},
	)

	grpcMux := runtime.NewServeMux(jsonOption)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := pb.RegisterSimpleBankHandlerServer(ctx, grpcMux, server)
	if err != nil {
		log.Fatal("cannot start gateway server:", err)
		return err
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	statikFS, err := fs.New()
	if err != nil {
		log.Fatal("cannot create statik fs", err)
		return err
	}
	swaggerHandler := http.StripPrefix("/swagger/", http.FileServer(statikFS))
	mux.Handle("/swagger/", swaggerHandler)

	listener, err := net.Listen("tcp", config.HTTPServerAddress)
	if err != nil {
		log.Fatal("cannot create listener:", err)
		return err
	}

	log.Println("starting HTTP gateway server on", config.HTTPServerAddress)
	err = http.Serve(listener, mux)
	if err != nil {
		log.Fatal("cannot start HTTP gateway server:", err)
		return err
	}

	return nil
}

func runHTTPServer(store persistence.Store, logger *slog.Logger, validate *validator.Validate, tokenMaker token.Maker, config util.Config) error {
	server := app.NewServer(store, logger, validate, tokenMaker, config)

	err := server.Start(config.HTTPServerAddress)
	if err != nil {
		log.Fatal("cannot start HTTP server:", err)
		return err
	}

	return nil
}
