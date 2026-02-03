package main

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lwandokasuba/golang-sqlc/internal/config"
	"github.com/lwandokasuba/golang-sqlc/internal/db"
	"github.com/lwandokasuba/golang-sqlc/internal/service"
	"github.com/lwandokasuba/golang-sqlc/internal/transport/http"
)

func main() {
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	connPool, err := pgxpool.New(context.Background(), cfg.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	store := db.NewStore(connPool)
	svc := service.NewService(store)
	server := http.NewServer(svc)

	err = server.Start(cfg.ServerAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}
