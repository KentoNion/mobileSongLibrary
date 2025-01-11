package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" //драйвер postgres
	goose "github.com/pressly/goose/v3"
	swagger "mobileSongLibrary/gates/apiservice"
	"mobileSongLibrary/gates/server"
	"mobileSongLibrary/gates/storage"
	"mobileSongLibrary/internal/config"
	"mobileSongLibrary/internal/logger"
	"net/http"
	"os"
)

//@title mobileSongLibrary
//@version 1.0.0
//@description API Server for songLibrary application

//@host sets with config
//@BasePath /

func main() {
	const op = "cmd.main"
	//Считываем конфиг
	cfg := config.MustLoad()

	//инициализируем логгер
	log := logger.MustInitLogger(cfg)
	log.Debug(op, "log", "logger started in debug mode")

	//Подключаемся к бд
	dbhost := os.Getenv("DB_HOST") //DB_HOST прописывается в docker_compose, если его там нет, значит считается из конфига
	if dbhost == "" {
		dbhost = cfg.DB.Host
	}
	connStr := fmt.Sprintf("user=%s password=%s dbname=mobile_song host=%s sslmode=%s timezone=UTC", cfg.DB.User, cfg.DB.Pass, dbhost, cfg.DB.Ssl)
	conn, err := sqlx.Connect("postgres", connStr) //подключение к бд
	if err != nil {
		panic(err)
	}
	db := storage.NewDB(conn, log) //переменная базы данных

	//накатываем миграцию
	migrationsPath := os.Getenv("MIGRATIONS_PATH") //для докера
	if migrationsPath == "" {
		migrationsPath = "./gates\\storage\\migrations"
	}

	//накатываем миграцию
	//err = goose.Down(conn.DB, migrationsPath)
	err = goose.Up(conn.DB, migrationsPath)
	if err != nil {
		panic(err)
	}
	//инициализируем сваггер
	restServerAddr := cfg.Rest.Host + ":" + cfg.Rest.Port
	client, err := swagger.NewClient(cfg.SwagCli.Host + ":" + cfg.SwagCli.Port)
	if err != nil {
		panic(err)
	}

	router := chi.NewRouter()
	_ = server.NewServer(router, db, log, client)

	log.Info("Starting server at port: " + cfg.Rest.Port)
	err = http.ListenAndServe(restServerAddr, router)
	if err != nil {
		panic(err)
		return
	}
}
