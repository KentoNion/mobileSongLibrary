package main

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"mobileSongLibrary/gates/postgres"
	"sync"
)

func main() {
	log, err := zap.NewDevelopment() // инструмент логирования ошибок
	if err != nil {
		panic(err)
	}

	ctx := context.Background() // контекст

	conn, err := sqlx.Connect("postgres", "songs.bd") //подключение к бд
	if err != nil {
		panic(err)
	}
	db := postgres.NewDB(conn) //переменная базы данных

	wg := sync.WaitGroup{} //wait group для синхронизации горутин

	router := chi.NewRouter() //роутер

}
