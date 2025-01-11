package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
	"log/slog"
	_ "mobileSongLibrary/docs"
	"mobileSongLibrary/domain"
	swagger "mobileSongLibrary/gates/apiservice"
	"mobileSongLibrary/gates/storage"
	"mobileSongLibrary/internal/config"
	"net/http"
	"strconv"
	"strings"
)

type Server struct {
	db      *storage.DB
	context context.Context
	log     *slog.Logger
	client  swagger.ClientInterface
	cfg     *config.Config
}

type groupRename struct {
	OldName string `json:"old_name"`
	NewName string `json:"new_name"`
}

type SongsStorage interface {
	AddSong(song storage.Song) error
	UpdateSong(song storage.Song) error
	GetSong(group domain.GroupName, songName domain.SongName) (domain.Song, error)
	DeleteSong(group domain.GroupName, song domain.SongName) error
	GetLibrary(ctx context.Context, filter domain.SongFilter) ([]domain.Song, error)
}

func NewServer(router *chi.Mux, db *storage.DB, log *slog.Logger, client swagger.ClientInterface, conf *config.Config) *Server {
	const op = "gates.Server.NewServer"
	server := &Server{
		db:      db,
		context: context.Background(),
		log:     log,
		client:  client,
		cfg:     conf,
	}

	router.Method(http.MethodGet, "/library", http.HandlerFunc(server.GetLibraryHandler))        //Хендлер на получение всей библиотеки песен
	router.Method(http.MethodGet, "/song", http.HandlerFunc(server.GetSongHandler))              //хендлер на получение конкретной песни
	router.Method(http.MethodDelete, "/song", http.HandlerFunc(server.DeleteSongHandler))        //Хендлер на удаление конкретной песни
	router.Method(http.MethodPost, "/song", http.HandlerFunc(server.AddSongHandler))             //хендлер на добавление новой песни
	router.Method(http.MethodPatch, "/song", http.HandlerFunc(server.UpdateSongHandler))         //Хендлер на изменение данных песни
	router.Method(http.MethodPatch, "/renamegroup", http.HandlerFunc(server.RenameGroupHandler)) //Хендлер на изменение название группы
	router.Method(http.MethodGet, "/info", http.HandlerFunc(server.InfoHandler))
	//swagger
	router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"), // Укажите базовый путь
	))
	server.log.Info(op, "router configured", "")
	return server
}

// AddSongHandler godoc
//
// @Summary      Добавить новую песню
// @Description  Добавляет новую песню в библиотеку
// @Tags         Songs
// @Accept       json
// @Produce      json
// @Param        song  body  domain.Song  true  "Данные новой песни"
// @Success      201     {string}  string  "Песня успешно добавлена"
// @Failure      400     {object}  string  "Некорректный запрос"
// @Failure      500     {object}  string  "Ошибка сервера"
// @Router       /song [post]
func (s Server) AddSongHandler(w http.ResponseWriter, r *http.Request) {
	const op = "gates.server.AddSongHandler"

	s.log.Info(op, "starting addSongHandler", "")
	var song domain.Song
	//читаем запрос
	if err := json.NewDecoder(r.Body).Decode(&song); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		s.log.Error(op, "failed to decode request body", "")
		return
	}
	err := song.Validate() //проверка на не пустые параметры group и song
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		s.log.Error(op, "failed to validate request body", "")
		return
	}

	//реализация API
	response, err := s.client.GetInfo(r.Context(), &swagger.GetInfoParams{string(song.GroupName), string(song.SongName)})
	fmt.Println(string(song.GroupName))
	fmt.Println(string(song.SongName))
	s.log.Debug(op, "Request parameters: group=", string(song.GroupName), "song=", string(song.SongName))

	if err != nil {
		s.log.Error(op, "failed to get info: ", err)
		http.Error(w, "Failed to get info: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer response.Body.Close()
	var songDetail swagger.SongDetail
	json.NewDecoder(response.Body).Decode(&songDetail)

	song.Link = domain.Link(songDetail.Link)
	song.Text = songDetail.Text
	song.ReleaseDate, err = domain.ParseCustomDate(songDetail.ReleaseDate)
	if err != nil {

	}

	defer r.Body.Close()
	if response.StatusCode != http.StatusOK {
		http.Error(w, "Unexpected status code: "+response.Status, http.StatusInternalServerError)
		s.log.Error(op, "got unexpected status code: ", response.Status)
		return
	}
	//пакуем песню в бд
	err = s.db.AddSong(storage.ToStorage(song))
	if err != nil {
		s.log.Error("Failed to add song", err)
		http.Error(w, "Failed to add song", http.StatusInternalServerError)
		return
	}
	//всё ок
	s.log.Info("AddSongHandler: successfully added song")
	w.WriteHeader(http.StatusCreated)
}

// обновит все старые данные на новые если строка не будет пустой (кроме имени группы и названии песни, он в renameGroupHandler),
// а имя песни изменить никак нельзя, песни вроде как не меняют имена, верно?

// UpdateSongHandler godoc
//
// @Summary      Обновить информацию о песне
// @Description  Обновляет данные о песне, кроме её названия
// @Tags         Songs
// @Accept       json
// @Produce      json
// @Param        song  body  domain.Song  true  "Обновлённые данные песни"
// @Success      200     {string}  string  "Песня успешно обновлена"
// @Failure      400     {object}  string  "Некорректный запрос"
// @Failure      500     {object}  string  "Ошибка сервера"
// @Router       /song [patch]
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

// GetLibraryHandler godoc
//
// @Summary      Получить всю библиотеку песен
// @Description  Возвращает список всех песен с возможностью фильтрации
// @Tags         Library
// @Accept       json
// @Produce      json
// @Param        filter  body  domain.SongFilter  true  "Фильтр для поиска песен"
// @Success      200     {array}  domain.Song
// @Failure      400     {object}  string  "Некорректный запрос"
// @Failure      500     {object}  string  "Ошибка сервера"
// @Router       /library [post]
func (s Server) GetLibraryHandler(w http.ResponseWriter, r *http.Request) {
	const op = "gates.Server.GetLibraryHandler"

	s.log.Info(op, "connected to GetLibraryHandler", "trying to get library")

	// Извлекаем параметры из строки запроса
	query := r.URL.Query()
	filter := domain.SongFilter{
		GroupName: query.Get("group"),
		SongName:  query.Get("song"),
	}

	if limit := query.Get("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			filter.Limit = l
		}
	}

	if offset := query.Get("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil {
			filter.Offset = o
		}
	}

	s.log.Debug(op, "filter:", filter)

	// Получаем библиотеку
	library, err := s.db.GetLibrary(s.context, filter)
	if err != nil {
		http.Error(w, "Failed to retrieve library: "+err.Error(), http.StatusInternalServerError)
		s.log.Error(op, "failed to retrieve library", err)
		return
	}

	// Возвращаем результат
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(library)
	s.log.Info(op, "successfully retrieved library")
}

// GetSongHandler godoc
//
// @Summary      Получить информацию о песне
// @Description  Возвращает данные о песне с пагинацией текста
// @Tags         Songs
// @Accept       json
// @Produce      json
// @Param        song  body  domain.Song  true  "Название группы и песни"
// @Success      200     {object}  map[string]interface{}
// @Failure      400     {object}  string  "Некорректный запрос"
// @Failure      500     {object}  string  "Ошибка сервера"
// @Router       /song [post]
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

// DeleteSongHandler godoc
//
// @Summary      Удалить песню
// @Description  Удаляет песню по названию и группе
// @Tags         Songs
// @Accept       json
// @Produce      json
// @Param        song  body  domain.Song  true  "Название группы и песни для удаления"
// @Success      200     {string}  string  "Успешное удаление"
// @Failure      400     {object}  string  "Некорректный запрос"
// @Failure      500     {object}  string  "Ошибка сервера"
// @Router       /song [delete]
func (s Server) DeleteSongHandler(w http.ResponseWriter, r *http.Request) {
	const op = "gates.Server.DeleteSongHandler"

	s.log.Info(op, "connected to DeleteSongHandler", "trying to delete song")
	var song domain.Song
	//читаем запрос
	if err := json.NewDecoder(r.Body).Decode(&song); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		s.log.Error(op, "failed to decode song", err)
		return
	}
	defer r.Body.Close()

	s.log.Debug(op, "got song: ", song)
	err := song.Validate() //проверка на не пустые параметры group и song
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		s.log.Error(op, "failed to validate song", err)
		return
	}

	//удаляем песню из бд
	err = s.db.DeleteSong(song.GroupName, song.SongName)
	if err != nil {
		http.Error(w, "Failed to delete song: "+err.Error(), http.StatusInternalServerError)
		s.log.Error(op, "failed to delete song", err)
		return
	}
	//пишем что всё ок
	s.log.Info("DeleteSongHandler: successfully deleted song")
	w.WriteHeader(http.StatusOK)
}

// RenameGroupHandler godoc
//
// @Summary      Переименовать группу
// @Description  Изменяет название музыкальной группы
// @Tags         Groups
// @Accept       json
// @Produce      json
// @Param        groupRename  body  groupRename  true  "Старое и новое название группы"
// @Success      204     {string}  string  "Группа успешно переименована"
// @Failure      400     {object}  string  "Некорректный запрос"
// @Failure      500     {object}  string  "Ошибка сервера"
// @Router       /renamegroup [patch]
func (s Server) RenameGroupHandler(w http.ResponseWriter, r *http.Request) {
	const op = "gates.Server.RenameGroupHandler"

	s.log.Info(op, "connected to RenameGroupHandler", "trying to rename group")
	var group groupRename

	//читаем запрос
	err := json.NewDecoder(r.Body).Decode(&group)
	if err != nil {

		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		s.log.Error(op, "failed to decode group", err)
		return
	}
	defer r.Body.Close()
	s.log.Debug(op, "old name", group.OldName, "new name", group.NewName)
	//переименовываем группу
	err = s.db.GroupRename(group.OldName, group.NewName)
	if err != nil {
		http.Error(w, "Failed to rename song: "+err.Error(), http.StatusInternalServerError)
		s.log.Error(op, "failed to rename song", err)
		return
	}
	//всё ок
	s.log.Info("RenameGroupHandler: successfully renamed song")
	w.WriteHeader(http.StatusNoContent)
}

func (s Server) InfoHandler(w http.ResponseWriter, r *http.Request) {
	const op = "gates.Server.InfoHandler"

	// Извлекаем параметры из запроса
	group := r.URL.Query().Get("group")
	song := r.URL.Query().Get("song")

	if group == "" || song == "" {
		http.Error(w, "Missing required parameters: group and song", http.StatusBadRequest)
		s.log.Error(op, "missing required parameters")
		return
	}

	// Логируем параметры
	s.log.Info(op, "Received request for info", "group", group, "song", song)

	// Ответ с заглушкой
	response := swagger.SongDetail{
		Link:        "https://www.youtube.com/watch?v=Xsp3_a-PMTw",
		ReleaseDate: "16.07.2006",
		Text:        "Ooh baby, don't you know I suffer?\nOoh baby, can you hear me moan?",
	}

	// Возвращаем JSON-ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.log.Error(op, "failed to write response", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
