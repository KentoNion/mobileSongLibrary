package main

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" //драйвер postgres
	"go.uber.org/zap"
	server "mobileSongLibrary/gates"
	"mobileSongLibrary/gates/postgres"
	"net/http"
)

func main() {
	log, err := zap.NewDevelopment() // инструмент логирования ошибок
	if err != nil {
		panic(err)
	}

	ctx := context.Background() // контекст

	postgres.RunGooseMigrations("songs")
	log.Info("Songs migrations applied successfully")

	conn, err := sqlx.Connect("postgres", "user=postgres password=postgres dbname=songs host=localhost sslmode=disable") //подключение к бд
	if err != nil {
		panic(err)
	}
	db := postgres.NewDB(conn) //переменная базы данных

	router := chi.NewRouter()
	_ = server.NewServer(ctx, router, db, log)

	log.Info("Starting server")
	err = http.ListenAndServe(":8080", router)
	if err != nil {
		log.Error("server error", zap.Error(err))
		return
	}
}
