package main

import (
	"log/slog"
	"os"

	"github.com/4aykovksi/medods_test_task/internal/config"
	"github.com/4aykovksi/medods_test_task/internal/repository/mongorepos"
	"github.com/4aykovksi/medods_test_task/internal/services"
	"github.com/4aykovksi/medods_test_task/pkg/database/mongodb"
	"github.com/4aykovksi/medods_test_task/pkg/lib/auth"
	"github.com/4aykovksi/medods_test_task/pkg/lib/hasher"
)

func main() {
	// parse config
	cfg := config.MustLoad()

	// init logger
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// init mondodb client
	mongoClient, err := mongodb.NewClient(cfg.Mongodb.URI)
	if err != nil {
		log.Error("can't init mongodb client", slog.String("err", err.Error()))
		os.Exit(1)
	}

	// connect to database
	db := mongoClient.Database(cfg.Mongodb.Database)

	// init repos
	userRepo := mongorepos.NewUserRepository(db)
	sessionRepo := mongorepos.NewRefreshSessionsRepository(db)

	// init services

	bcryptHasher := hasher.NewBcryptHasher()
	tokenManager := auth.NewManager(cfg.Secret)

	sessionService := services.NewRefreshSessionService(sessionRepo, bcryptHasher)
	userService := services.NewUserService(userRepo, sessionService, tokenManager, bcryptHasher, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)

	// TODO: init router
	// TODO: run server
}
