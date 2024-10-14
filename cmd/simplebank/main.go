package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/RobinHood3082/simplebank/internal/app"
	"github.com/RobinHood3082/simplebank/internal/gapi"
	"github.com/RobinHood3082/simplebank/internal/pb"
	"github.com/RobinHood3082/simplebank/internal/persistence"
	"github.com/RobinHood3082/simplebank/internal/token"
	"github.com/RobinHood3082/simplebank/mail"
	"github.com/RobinHood3082/simplebank/util"
	"github.com/RobinHood3082/simplebank/worker"
	"github.com/go-playground/validator/v10"
	"github.com/golang-migrate/migrate/v4"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rakyll/statik/fs"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"

	_ "github.com/RobinHood3082/simplebank/doc/statik"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

var interruptSignals = []os.Signal{
	os.Interrupt,
	syscall.SIGTERM,
	syscall.SIGINT,
}

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config:", err)
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), interruptSignals...)
	defer stop()

	conn, err := pgxpool.New(ctx, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
		return
	}

	redisOpt := asynq.RedisClientOpt{
		Addr: config.RedisAddress,
	}

	taskDistributor := worker.NewRedisTaskDistributor(redisOpt)

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

	waitGroup, ctx := errgroup.WithContext(ctx)

	runTaskProcessor(ctx, waitGroup, redisOpt, store, config)
	runGatewayServer(ctx, waitGroup, store, logger, tokenMaker, config, taskDistributor)
	runGRPCServer(ctx, waitGroup, store, logger, tokenMaker, config, taskDistributor)

	err = waitGroup.Wait()
	if err != nil {
		log.Fatal("error from wait group:", err)
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

func runTaskProcessor(
	ctx context.Context,
	waitGroup *errgroup.Group,
	redisOpt asynq.RedisClientOpt,
	store persistence.Store,
	config util.Config,
) {
	mailer := mail.NewGmailSender(
		config.EmailSenderName,
		config.EmailSenderAddress,
		config.EmailSenderPassword,
	)
	taskProcessor := worker.NewRedisTaskProcessor(redisOpt, store, mailer)
	log.Println("task processor starting")

	err := taskProcessor.Start()
	if err != nil {
		log.Fatal("failed to start task processor:", err)
	}

	waitGroup.Go(
		func() error {
			<-ctx.Done()
			log.Println("task processor shutting down")
			taskProcessor.Shutdown()
			log.Println("task processor stopped")
			return nil
		},
	)
}

func runGRPCServer(
	ctx context.Context,
	waitGroup *errgroup.Group,
	store persistence.Store,
	logger *slog.Logger,
	tokenMaker token.Maker,
	config util.Config,
	taskDistributor worker.TaskDistributor,
) {
	server := gapi.NewServer(store, logger, tokenMaker, config, taskDistributor)

	grpcServer := grpc.NewServer()
	pb.RegisterSimpleBankServer(grpcServer, server)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", config.GRPCServerAddress)
	if err != nil {
		log.Fatal("cannot start gRPC server:", err)
	}

	waitGroup.Go(
		func() error {
			log.Println("starting gRPC server on", config.GRPCServerAddress)

			err = grpcServer.Serve(listener)
			if err != nil {
				if errors.Is(err, grpc.ErrServerStopped) {
					return nil
				}

				log.Fatal("gRPC server failed to serve")
				return err
			}
			return nil
		},
	)

	waitGroup.Go(
		func() error {
			<-ctx.Done()
			log.Println("gracefully shutting down gRPC server")

			grpcServer.GracefulStop()
			log.Println("gRPC server stopped")
			return nil
		},
	)
}

func runGatewayServer(
	ctx context.Context,
	waitGroup *errgroup.Group,
	store persistence.Store,
	logger *slog.Logger,
	tokenMaker token.Maker,
	config util.Config,
	taskDistributor worker.TaskDistributor,
) {
	server := gapi.NewServer(store, logger, tokenMaker, config, taskDistributor)

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

	err := pb.RegisterSimpleBankHandlerServer(ctx, grpcMux, server)
	if err != nil {
		log.Fatal("cannot start gateway server:", err)
		return
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	statikFS, err := fs.New()
	if err != nil {
		log.Fatal("cannot create statik fs", err)
	}

	swaggerHandler := http.StripPrefix("/swagger/", http.FileServer(statikFS))
	mux.Handle("/swagger/", swaggerHandler)

	httpServer := &http.Server{
		Handler: mux,
		Addr:    config.HTTPServerAddress,
	}

	waitGroup.Go(
		func() error {
			log.Println("starting HTTP gateway server on", httpServer.Addr)
			err = httpServer.ListenAndServe()
			if err != nil {
				if errors.Is(err, http.ErrServerClosed) {
					return nil
				}
				log.Fatal("HTTP gateway server failed to serve")
				return err
			}

			return nil
		},
	)

	waitGroup.Go(
		func() error {
			<-ctx.Done()
			log.Println("gracefully shutting down HTTP gateway server")

			err = httpServer.Shutdown(context.Background())
			if err != nil {
				log.Fatal("cannot shutdown HTTP gateway server")
				return err
			}

			log.Println("HTTP gateway server stopped")
			return nil
		},
	)
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
