package domain

import (
	"net/http"
)

type DB interface { //интерфейс для работы с БД
	GetLibraryHandler(http.ResponseWriter, *http.Request)
	GetSongHandler(http.ResponseWriter, *http.Request)
	DeleteSongHandler(http.ResponseWriter, *http.Request)
	addSongHandler(http.ResponseWriter, *http.Request)
}
