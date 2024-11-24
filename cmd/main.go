package main

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq" //драйвер postgres
	"go.uber.org/zap"

	"mobileSongLibrary/gates/postgres"
	"mobileSongLibrary/gates/server"
	"mobileSongLibrary/internal/logger"
	"net/http"
	"os"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}

	log, err := logger.InitLogger(os.Getenv("LOG_FILE")) // инструмент логирования ошибок
	if err != nil {
		panic(err)
	}

	ctx := context.Background() // контекст

	postgres.RunGooseMigrations(os.Getenv("DB_NAME"))
	log.Info("Songs migrations applied successfully")

	conn, err := sqlx.Connect("postgres", fmt.Sprintf("user=%v password=%v dbname=%v host=%v sslmode=%v", os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"), os.Getenv("DB_HOST"), os.Getenv("DB_SSLMODE"))) //подключение к бд
	if err != nil {
		panic(err)
	}
	db := postgres.NewDB(conn) //переменная базы данных

	router := chi.NewRouter()
	_ = server.NewServer(ctx, router, db, log)

	log.Info("Starting server at port: " + os.Getenv("SERVER_PORT"))
	err = http.ListenAndServe(os.Getenv("SERVER_HOST")+":"+os.Getenv("SERVER_PORT"), router)
	if err != nil {
		log.Error("server error", zap.Error(err))
		return
	}
}
