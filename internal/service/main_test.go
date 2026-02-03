package service

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lwandokasuba/golang-sqlc/internal/config"
	"github.com/lwandokasuba/golang-sqlc/internal/db"
)

var testStore db.Store

func TestMain(m *testing.M) {
	config, err := config.LoadConfig("../..")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	connPool, err := pgxpool.New(context.Background(), config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	testStore = db.NewStore(connPool)

	os.Exit(m.Run())
}
