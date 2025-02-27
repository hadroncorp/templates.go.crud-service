package main

import (
	"log/slog"

	"github.com/caarlos0/env/v11"
	"github.com/doug-martin/goqu/v9"
	"github.com/hadroncorp/geck/persistence/driver/postgres"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		slog.Warn("cannot load .env file, using OS environment variables", slog.String("err", err.Error()))
	}

	postgresConfig, err := env.ParseAs[postgres.DBConfig]()
	if err != nil {
		panic(err)
	}
	db, err := postgres.NewPooledDB(postgresConfig)
	if err != nil {
		panic(err)
	}

	goqu.Insert("platform_users").Rows().Executor()
}
