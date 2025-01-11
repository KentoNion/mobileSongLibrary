package server

import (
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"log/slog"
	"mobileSongLibrary/domain"
	swagger "mobileSongLibrary/gates/apiservice"
	"mobileSongLibrary/gates/storage"
	"net/http"
	"strconv"
	"strings"
)

type Server struct {
	db      *storage.DB
	context context.Context
	log     *slog.Logger
	client  swagger.ClientInterface
}

type groupRename struct {
	oldName string
	newName string
}

type SongsStorage interface {
	AddSong(song storage.Song) error
	UpdateSong(song storage.Song) error
	GetSong(group domain.GroupName, songName domain.SongName) (domain.Song, error)
	DeleteSong(group domain.GroupName, song domain.SongName) error
	GetLibrary(ctx context.Context, filter domain.SongFilter) ([]domain.Song, error)
}

func NewServer(router *chi.Mux, db *storage.DB, log *slog.Logger, client swagger.ClientInterface) *Server {
	const op = "gates.Server.NewServer"
	server := &Server{
		db:      db,
		context: context.Background(),
		log:     log,
		client:  client,
	}

	router.Method(http.MethodGet, "/library", http.HandlerFunc(server.GetLibraryHandler))      //Хендлер на получение всей библиотеки песен
	router.Method(http.MethodGet, "/song", http.HandlerFunc(server.GetSongHandler))            //хендлер на получение конкретной песни
	router.Method(http.MethodDelete, "/song", http.HandlerFunc(server.DeleteSongHandler))      //Хендлер на удаление конкретной песни
	router.Method(http.MethodPost, "/song", http.HandlerFunc(server.AddSongHandler))           //хендлер на добавление новой песни
	router.Method(http.MethodPut, "/song", http.HandlerFunc(server.UpdateSongHandler))         //Хендлер на изменение данных песни
	router.Method(http.MethodPut, "/renamegroup", http.HandlerFunc(server.RenameGroupHandler)) //Хендлер на изменение название группы

	//swagger
	//router.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL("http://localhost:8080/swagger/doc.json")))
	server.log.Info(op, "router configured", "")
	return server
}

func (s Server) AddSongHandler(w http.ResponseWriter, r *http.Request) {
	const op = "gates.Server.AddSongHandler"

	s.log.Info(op, "connected to AddSongHandler", "trying to add song")
	var song domain.Song
	//читаем запрос
	if err := json.NewDecoder(r.Body).Decode(&song); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		s.log.Error(op, "failed to decode song", err)
		return
	}
	err := song.Validate() //проверка на не пустые параметры group и song
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		s.log.Error(op, "failed to validate song", err)
	}

	//реализация API
	response, err := s.client.GetInfo(r.Context(), &swagger.GetInfoParams{string(song.GroupName), string(song.SongName)})
	if err != nil {
		s.log.Error(op, "failed to get info", err)
		http.Error(w, "Failed to get info: "+err.Error(), http.StatusInternalServerError)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		http.Error(w, "Unexpected status code: "+response.Status, http.StatusInternalServerError)
		return
	}

	var songDetail swagger.SongDetail
	json.NewDecoder(response.Body).Decode(&songDetail)

	song.Link = domain.Link(songDetail.Link)
	song.Text = songDetail.Text
	song.ReleaseDate, err = domain.ParseCustomDate(songDetail.ReleaseDate)
	if err != nil {
		s.log.Error(op, "failed to parse release date", err)
		http.Error(w, "Failed to get release date: "+err.Error(), http.StatusInternalServerError)
	}

	defer r.Body.Close()
	//пакуем песню в бд
	err = s.db.AddSong(storage.ToStorage(song))
	if err != nil {
		s.log.Error(op, "failed to add song", err)
		http.Error(w, "Failed to add song", http.StatusInternalServerError)
		return
	}
	//всё ок
	s.log.Info(op, "song added successfully", song.SongName)
	w.WriteHeader(http.StatusCreated)
}

// обновит все старые данные на новые если строка не будет пустой (кроме имени группы и названии песни, он в renameGroupHandler),
// а имя песни изменить никак нельзя, песни вроде как не меняют имена, верно?
func (s Server) UpdateSongHandler(w http.ResponseWriter, r *http.Request) {
	const op = "gates.Server.UpdateSongHandler"

	s.log.Info(op)
	var song domain.Song
	//читаем запрос
	if err := json.NewDecoder(r.Body).Decode(&song); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		s.log.Error(op, "failed to decode song", err)
		return
	}
	defer r.Body.Close()

	err := song.Validate() //проверка на не пустые параметры group и song
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		s.log.Error(op, "failed to validate song", err)
	}

	//обновляем песню
	err = s.db.UpdateSong(storage.ToStorage(song))
	if err != nil {
		s.log.Error(op, "failed to update song", err)
		http.Error(w, "Failed to update song", http.StatusInternalServerError)
		return
	}
	//всё ок
	s.log.Info("UpdateSongHandler: successfully updated song")
	w.WriteHeader(http.StatusCreated)
}

func (s Server) GetLibraryHandler(w http.ResponseWriter, r *http.Request) {
	const op = "gates.Server.GetLibraryHandler"

	s.log.Info(op, "connected to GetLibraryHandler", "trying to get library")
	var filter domain.SongFilter
	//читаем запрос
	if err := json.NewDecoder(r.Body).Decode(&filter); err != nil { //пагинация реализованна в фильтре
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		s.log.Error(op, "failed to decode filter", err)
		return
	}
	s.log.Debug(op, "got filter: ", filter)
	defer r.Body.Close()
	//достаём библиотеку наших хитов из бд
	library, err := s.db.GetLibrary(s.context, filter)
	if err != nil {
		http.Error(w, "Failed to retrieve library: "+err.Error(), http.StatusInternalServerError)
		s.log.Error(op, "failed to retrieve library", err)
		return
	}
	//формируем ответ в Джейсона
	response, err := json.Marshal(library)
	if err != nil {
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
		s.log.Error(op, "failed to encode response", err)
		return
	}
	//загружаем в интернеты
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
	s.log.Info("GetLibraryHandler: successfully retrieved library")
}

func (s Server) GetSongHandler(w http.ResponseWriter, r *http.Request) {
	const op = "gates.Server.GetSongHandler"

	s.log.Info(op, "connected to GetSongHandler", "trying to get song")
	var song domain.Song
	//читаем запрос
	if err := json.NewDecoder(r.Body).Decode(&song); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		s.log.Error(op, "failed to decode song", err)
		return
	}
	defer r.Body.Close()

	err := song.Validate() //проверка на не пустые параметры group и song
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		s.log.Error(op, "failed to validate song", err)
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
	song, err = s.db.GetSong((song.GroupName), song.SongName)
	if err != nil {
		http.Error(w, "Failed to retrieve song: "+err.Error(), http.StatusInternalServerError)
		s.log.Error(op, "failed to retrieve song", err)
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
		s.log.Error(op, "failed to encode response", err)
	}
	w.Header().Set("Content-Type", "application/json") //загружаем
	w.WriteHeader(http.StatusOK)
	w.Write(response)
	s.log.Info("GetSongHandler: successfully retrieved song")
}

func (s Server) DeleteSongHandler(w http.ResponseWriter, r *http.Request) {
	const op = "gates.Server.DeleteSongHandler"

	s.log.Info(op, "connected to DeleteSongHandler", "trying to delete song")
	var song storage.Song
	//читаем запрос
	if err := json.NewDecoder(r.Body).Decode(&song); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		s.log.Error(op, "failed to decode song", err)
		return
	}
	defer r.Body.Close()

	err := song.Validate() //проверка на не пустые параметры group и song
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		s.log.Error(op, "failed to validate song", err)
	}

	//удаляем песню из бд
	err = s.db.DeleteSong(song.GroupName, song.SongName)
	if err != nil {
		http.Error(w, "Failed to delete song: "+err.Error(), http.StatusInternalServerError)
		s.log.Error(op, "failed to delete song", err)
	}
	//пишем что всё ок
	s.log.Info("DeleteSongHandler: successfully deleted song")
	w.WriteHeader(http.StatusOK)
}

func (s Server) RenameGroupHandler(w http.ResponseWriter, r *http.Request) {
	const op = "gates.Server.RenameGroupHandler"

	s.log.Info(op, "connected to RenameGroupHandler", "trying to rename group")
	var group groupRename
	//читаем запрос
	if err := json.NewDecoder(r.Body).Decode(&group); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		s.log.Error(op, "failed to decode group", err)
		return
	}
	defer r.Body.Close()
	//переименовываем группу
	err := s.db.GroupRename(group.oldName, group.newName)
	if err != nil {
		http.Error(w, "Failed to rename song: "+err.Error(), http.StatusInternalServerError)
		s.log.Error(op, "failed to rename song", err)
		return
	}
	//всё ок
	s.log.Info("RenameGroupHandler: successfully renamed song")
	w.WriteHeader(http.StatusNoContent)
}
