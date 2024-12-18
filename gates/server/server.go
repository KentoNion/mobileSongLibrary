package server

import (
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"mobileSongLibrary/domain"
	swagger "mobileSongLibrary/gates/apiservice"
	"mobileSongLibrary/gates/postgres"
	"net/http"
	"strconv"
	"strings"
)

type Server struct {
	db      *postgres.DB
	context context.Context
	log     *zap.Logger
	client  swagger.ClientInterface
}

type groupRename struct {
	oldName string
	newName string
}

func NewServer(ctx context.Context, router *chi.Mux, db *postgres.DB, log *zap.Logger, client swagger.ClientInterface) *Server {
	server := &Server{
		db:      db,
		context: ctx,
		log:     log,
		client:  client,
	}

	router.Method(http.MethodGet, "/library", http.HandlerFunc(server.GetLibraryHandler))      //Хендлер на получение всей библиотеки песен
	router.Method(http.MethodGet, "song", http.HandlerFunc(server.GetSongHandler))             //хендлер на получение конкретной песни
	router.Method(http.MethodDelete, "/song", http.HandlerFunc(server.DeleteSongHandler))      //Хендлер на удаление конкретной песни
	router.Method(http.MethodPost, "/song", http.HandlerFunc(server.AddSongHandler))           //хендлер на добавление новой песни
	router.Method(http.MethodPut, "/song", http.HandlerFunc(server.UpdateSongHandler))         //Хендлер на изменение данных песни                                                         //router.HandleFunc("/updatesong", server.UpdateSongHandler)
	router.Method(http.MethodPut, "/renamegroup", http.HandlerFunc(server.RenameGroupHandler)) //Хендлер на изменение название группы

	server.log.Info("router configured")
	return server
}

func (s Server) AddSongHandler(w http.ResponseWriter, r *http.Request) {
	s.log.Info("AddSongHandler: connected to AddSongHandler", zap.String("method", r.Method), zap.String("path", r.URL.Path))
	var song domain.Song
	//читаем запрос
	if err := json.NewDecoder(r.Body).Decode(&song); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		s.log.Error("Failed to decode request body", zap.Error(err))
		return
	}
	err := song.Validate() //проверка на не пустые параметры group и song
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		s.log.Error("Invalid request body", zap.String("method", r.Method), zap.String("path", r.URL.Path))
	}

	//реализация API
	response, err := s.client.GetInfo(r.Context(), &swagger.GetInfoParams{song.GroupName, song.SongName})
	if err != nil {
		s.log.Error("Failed to get info", zap.String("method", r.Method), zap.String("path", r.URL.Path))
		http.Error(w, "Failed to get info: "+err.Error(), http.StatusInternalServerError)
	}
	defer response.Body.Close()
	var songDetail swagger.SongDetail
	json.NewDecoder(response.Body).Decode(&songDetail)

	song.Link = songDetail.Link
	song.Text = songDetail.Text
	song.ReleaseDate = songDetail.ReleaseDate

	defer r.Body.Close()
	//пакуем песню в бд
	err = s.db.AddSong(song)
	if err != nil {
		s.log.Error("Failed to add song", zap.Error(err))
		http.Error(w, "Failed to add song", http.StatusInternalServerError)
		return
	}
	//всё ок
	s.log.Info("AddSongHandler: successfully added song")
	w.WriteHeader(http.StatusCreated)
}

// обновит все старые данные на новые если строка не будет пустой (кроме имени группы и названии песни, он в renameGroupHandler),
// а имя песни изменить никак нельзя, песни вроде как не меняют имена, верно?
func (s Server) UpdateSongHandler(w http.ResponseWriter, r *http.Request) {
	s.log.Info("UpdateSongHandler: connected to UpdateSongHandler", zap.String("method", r.Method), zap.String("path", r.URL.Path))
	var song domain.Song
	//читаем запрос
	if err := json.NewDecoder(r.Body).Decode(&song); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		s.log.Error("Failed to decode request body", zap.Error(err))
		return
	}
	defer r.Body.Close()

	err := song.Validate() //проверка на не пустые параметры group и song
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		s.log.Error("Invalid request body", zap.String("method", r.Method), zap.String("path", r.URL.Path))
	}

	//обновляем песню
	err = s.db.UpdateSong(song)
	if err != nil {
		s.log.Error("Failed to update song", zap.Error(err))
		http.Error(w, "Failed to update song", http.StatusInternalServerError)
		return
	}
	//всё ок
	s.log.Info("UpdateSongHandler: successfully updated song")
	w.WriteHeader(http.StatusCreated)
}

func (s Server) GetLibraryHandler(w http.ResponseWriter, r *http.Request) {
	s.log.Info("GetLibraryHandler: connected to GetLibraryHandler", zap.String("method", r.Method))
	var filter domain.SongFilter
	//читаем запрос
	if err := json.NewDecoder(r.Body).Decode(&filter); err != nil { //пагинация реализованна в фильтре
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		s.log.Error("Failed to decode request body", zap.Error(err))
		return
	}
	defer r.Body.Close()
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
	s.log.Info("GetLibraryHandler: successfully retrieved library")
}

func (s Server) GetSongHandler(w http.ResponseWriter, r *http.Request) {
	s.log.Info("GetSongHandler: connected to GetSongHandler", zap.String("method", r.Method))
	var song postgres.Song
	//читаем запрос
	if err := json.NewDecoder(r.Body).Decode(&song); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		s.log.Error("Failed to decode request body", zap.Error(err))
		return
	}
	defer r.Body.Close()

	err := song.Validate() //проверка на не пустые параметры group и song
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		s.log.Error("Invalid request body", zap.String("method", r.Method), zap.String("path", r.URL.Path))
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
	song, err = s.db.GetSong(song.GroupName, song.SongName)
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
		"group":           song.GroupName,
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
	s.log.Info("GetSongHandler: successfully retrieved song")
}

func (s Server) DeleteSongHandler(w http.ResponseWriter, r *http.Request) {
	s.log.Info("DeleteSongHandler: connected to DeleteSongHandler", zap.String("method", r.Method))
	var song postgres.Song
	//читаем запрос
	if err := json.NewDecoder(r.Body).Decode(&song); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		s.log.Error("Failed to decode request body", zap.Error(err))
		return
	}
	defer r.Body.Close()

	err := song.Validate() //проверка на не пустые параметры group и song
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		s.log.Error("Invalid request body", zap.String("method", r.Method), zap.String("path", r.URL.Path))
	}

	//удаляем песню из бд
	err = s.db.DeleteSong(song.GroupName, song.SongName)
	if err != nil {
		http.Error(w, "Failed to delete song: "+err.Error(), http.StatusInternalServerError)
		s.log.Error("Failed to delete song", zap.Error(err))
	}
	//пишем что всё ок
	s.log.Info("DeleteSongHandler: successfully deleted song")
	w.WriteHeader(http.StatusNoContent) //todo так чтоль? или статус ок?
}

func (s Server) RenameGroupHandler(w http.ResponseWriter, r *http.Request) {
	s.log.Info("RenameGroupHandler: connected to RenameGroupHandler", zap.String("method", r.Method))
	var group groupRename
	//читаем запрос
	if err := json.NewDecoder(r.Body).Decode(&group); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		s.log.Error("Failed to decode request body", zap.Error(err))
		return
	}
	defer r.Body.Close()
	//переименовываем группу
	err := s.db.GroupRename(group.oldName, group.newName)
	if err != nil {
		http.Error(w, "Failed to rename song: "+err.Error(), http.StatusInternalServerError)
		s.log.Error("Failed to rename song", zap.Error(err))
		return
	}
	//всё ок
	s.log.Info("RenameGroupHandler: successfully renamed song")
	w.WriteHeader(http.StatusNoContent)
}
