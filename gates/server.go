package server

import (
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"mobileSongLibrary/gates/postgres"
	"net/http"
)

type Server struct {
	db *postgres.DB
	//логгер
}

func NewServer(ctx context.Context, router *chi.Mux, db *postgres.DB) *Server {
	server := &Server{
		db: db,
	}

	router.HandleFunc("/Library", server.GetLibraryHandler)
	return server
}

func (s Server) GetLibraryHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var filter postgres.SongFilter
	if err := json.NewDecoder(r.Body).Decode(&filter); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		zap.L().Error("Failed to decode request body", zap.Error(err))
		return
	}

	library, err := s.db.GetLibrary(ctx, filter)
	if err != nil {
		http.Error(w, "Failed to retrieve library: "+err.Error(), http.StatusInternalServerError)
		zap.L().Error("Failed to retrieve library", zap.Error(err))
		return
	}

	response, err := json.Marshal(library)
	if err != nil {
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
		zap.L().Error("Failed to encode response", zap.Error(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
