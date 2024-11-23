package server

import (
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"mobileSongLibrary/gates/postgres"
	"net/http"
	"strconv"
	"strings"
)

type Server struct {
	db      *postgres.DB
	context context.Context
	log     *zap.Logger
}

type groupRename struct {
	oldName string
	newName string
}

func NewServer(ctx context.Context, router *chi.Mux, db *postgres.DB, log *zap.Logger) *Server {
	server := &Server{
		db:      db,
		context: ctx,
		log:     log,
	}

	router.HandleFunc("/Library", server.GetLibraryHandler)
	router.HandleFunc("/song", server.GetSongHandler)
	router.HandleFunc("/deletesong", server.DeleteSongHandler)
	router.HandleFunc("/uploadsong", server.AddSongHandler)
	router.HandleFunc("/updatesong", server.UpdateSongHandler)
	router.HandleFunc("/renamegroup", server.RenameGroupHandler)

	return server
}

// Если попытаться добавить ту же песню что уже есть, то тогда она произведёт update старой версии и заменит старую информацию на предоставленую новую
func (s Server) AddSongHandler(w http.ResponseWriter, r *http.Request) {
	var song postgres.Song
	//читаем запрос
	if err := json.NewDecoder(r.Body).Decode(&song); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		s.log.Error("Failed to decode request body", zap.Error(err))
		return
	}
	//пакуем песню в бд
	err := s.db.AddSong(song)
	if err != nil {
		s.log.Error("Failed to add song", zap.Error(err))
		http.Error(w, "Failed to add song", http.StatusInternalServerError)
		return
	}
	//всё ок
	w.WriteHeader(http.StatusOK)
}

// обновит все старые данные на новые если строка не будет пустой (кроме имени группы и названии песни, он в renameGroupHandler),
// а имя песни изменить никак нельзя, песни вроде как не меняют имена, верно?
func (s Server) UpdateSongHandler(w http.ResponseWriter, r *http.Request) {
	var song postgres.Song
	//читаем запрос
	if err := json.NewDecoder(r.Body).Decode(&song); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		s.log.Error("Failed to decode request body", zap.Error(err))
		return
	}
	//обновляем песню
	err := s.db.UpdateSong(song)
	if err != nil {
		s.log.Error("Failed to update song", zap.Error(err))
		http.Error(w, "Failed to update song", http.StatusInternalServerError)
		return
	}
	//всё ок
	w.WriteHeader(http.StatusOK)
}

func (s Server) GetLibraryHandler(w http.ResponseWriter, r *http.Request) {
	var filter postgres.SongFilter //todo надо проверить работает ли фильтр
	//читаем запрос
	if err := json.NewDecoder(r.Body).Decode(&filter); err != nil { //пагинация реализованна в фильтре
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		s.log.Error("Failed to decode request body", zap.Error(err))
		return
	}
	//достаём библиотеку наших хитов из бд
	library, err := s.db.GetLibrary(s.context, filter)
	if err != nil {
		http.Error(w, "Failed to retrieve library: "+err.Error(), http.StatusInternalServerError)
		s.log.Error("Failed to retrieve library", zap.Error(err))
		return
	}
	//формируем ответ в Джейсона
	response, err := json.Marshal(library)
	if err != nil {
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
		s.log.Error("Failed to encode response", zap.Error(err))
		return
	}
	//загружаем в интернеты
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func (s Server) GetSongHandler(w http.ResponseWriter, r *http.Request) {
	var song postgres.Song
	//читаем запрос
	if err := json.NewDecoder(r.Body).Decode(&song); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		s.log.Error("Failed to decode request body", zap.Error(err))
		return
	}

	//Получение параметров пагинации из запроса
	query := r.URL.Query()
	page := 1
	size := 2
	if p := query.Get("page"); p != "" {
		if parsedPage, err := strconv.Atoi(p); err == nil {
			page = parsedPage
		}
	}
	if s := query.Get("size"); s != "" {
		if parsedSize, err := strconv.Atoi(s); err == nil {
			size = parsedSize
		}
	}
	//вытаскиваем песню из дб
	song, err := s.db.GetSong(song.Group, song.SongName)
	if err != nil {
		http.Error(w, "Failed to retrieve song: "+err.Error(), http.StatusInternalServerError)
		s.log.Error("Failed to retrieve song", zap.Error(err))
	}

	//пагинация
	verses := strings.Split(song.Text, "\n\n")
	start := (page - 1) * size
	end := start + size
	if start > len(verses) {
		start = len(verses)
	}
	if end > len(verses) {
		end = len(verses)
	}

	//формирование ответа
	resp := map[string]interface{}{
		"group":           song.Group,
		"song":            song.SongName,
		"release_date":    song.ReleaseDate,
		"link":            song.Link,
		"verses":          verses,
		"total_verses":    len(verses),
		"page":            page,
		"verses_per_page": size,
	}
	//пакуем ответ в джейсона
	response, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
		s.log.Error("Failed to encode response", zap.Error(err))
	}
	w.Header().Set("Content-Type", "application/json") //загружаем
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func (s Server) DeleteSongHandler(w http.ResponseWriter, r *http.Request) {
	var song postgres.Song
	//читаем запрос
	if err := json.NewDecoder(r.Body).Decode(&song); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		s.log.Error("Failed to decode request body", zap.Error(err))
		return
	}
	//удаляем песню из бд
	err := s.db.DeleteSong(song.Group, song.SongName)
	if err != nil {
		http.Error(w, "Failed to delete song: "+err.Error(), http.StatusInternalServerError)
		s.log.Error("Failed to delete song", zap.Error(err))
	}
	//пишем что всё ок
	w.WriteHeader(http.StatusNoContent) //todo так чтоль? или статус ок?
}

func (s Server) RenameGroupHandler(w http.ResponseWriter, r *http.Request) {
	var group groupRename
	//читаем запрос
	if err := json.NewDecoder(r.Body).Decode(&group); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		s.log.Error("Failed to decode request body", zap.Error(err))
		return
	}
	//переименовываем группу
	err := s.db.GroupRename(group.oldName, group.newName)
	if err != nil {
		http.Error(w, "Failed to rename song: "+err.Error(), http.StatusInternalServerError)
		s.log.Error("Failed to rename song", zap.Error(err))
		return
	}
	//всё ок
	w.WriteHeader(http.StatusOK)
}
