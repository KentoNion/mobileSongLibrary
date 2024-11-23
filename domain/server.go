package domain

import (
	"context"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func server(ctx context.Context) error {
	router := chi.NewRouter() //роутер
	router.Get("/getlibrary", getlibrary)
	router.Get("/getsong", getSong)
	router.Delete("/deletesong", deleteSong)
}

func getlibrary(w http.ResponseWriter, r *http.Request) {

}
