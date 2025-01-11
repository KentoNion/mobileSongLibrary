package storage

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" //драйвер postgres
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
	"mobileSongLibrary/domain"
	"mobileSongLibrary/internal/config"
	"mobileSongLibrary/internal/logger"
	"os"
	"testing"
	"time"
)

func TestInsertUpdateSelectGetLibraryRenameGroupDelete(t *testing.T) {
	ctx := context.Background()

	// Считываем конфиг
	os.Setenv("CONFIG_PATH", "../../../config.yaml")
	cfg := config.MustLoad()

	// Инициализация логгера
	log := logger.MustInitLogger(cfg)

	// Подключение к БД и накатывание миграций
	connStr := fmt.Sprintf(
		"user=%s password=%s dbname=mobile_song host=%s sslmode=%s timezone=UTC",
		cfg.DB.User, cfg.DB.Pass, cfg.DB.Host, cfg.DB.Ssl,
	)
	conn, err := sqlx.Connect("postgres", connStr)
	require.NoError(t, err)

	// Накатываем миграции
	err = goose.Up(conn.DB, "./migrations")
	require.NoError(t, err)
	t.Log("Test database migrations applied successfully")

	db := NewDB(conn, log)

	// Создание тестовых данных
	testSongs := []Song{
		{
			GroupName:   "muse",
			SongName:    "Supermassive Black Hole",
			ReleaseDate: time.Date(2006, time.July, 16, 0, 0, 0, 0, time.UTC),
			Text:        "Ooh baby, don't you know I suffer?",
			Link:        "https://www.youtube.com/watch?v=Xsp3_a-PMTw",
		},
		{
			GroupName:   "muse",
			SongName:    "WON'T STAND DOWN",
			ReleaseDate: time.Date(2022, time.January, 13, 0, 0, 0, 0, time.UTC),
			Text:        "I never believed that I would concede...",
			Link:        "https://youtu.be/d55ELY17CFM",
		},
		{
			GroupName:   "Buku",
			SongName:    "Front to Back",
			ReleaseDate: time.Date(2016, time.August, 30, 0, 0, 0, 0, time.UTC),
			Text:        "Front to the back...",
			Link:        "https://www.youtube.com/watch?v=PWROws51oWM",
		},
	}

	// Загружаем тестовые данные
	for _, testSong := range testSongs {
		err = db.AddSong(testSong)
		require.NoError(t, err)
	}

	// Тестируем обновление песни
	err = db.UpdateSong(Song{
		GroupName:   "Buku",
		SongName:    "Front to Back",
		ReleaseDate: time.Date(2006, time.July, 16, 0, 0, 0, 0, time.UTC),
	})
	require.NoError(t, err)

	songFromDB, err := db.GetSong("Buku", "Front to Back")
	require.NoError(t, err)
	require.Equal(t, time.Date(2006, time.July, 16, 0, 0, 0, 0, time.UTC), time.Time(songFromDB.ReleaseDate))

	// Тестируем переименование группы
	err = db.GroupRename("muse", "Muse")
	testSongs[0].GroupName = "Muse"
	testSongs[1].GroupName = "Muse"
	require.NoError(t, err)

	// Проверяем, что группа была переименована
	testLibrary, err := db.GetLibrary(ctx, domain.SongFilter{GroupName: "Muse"})
	require.NoError(t, err)
	require.Len(t, testLibrary, 2)

	// Удаляем данные
	for _, song := range testSongs {
		err = db.DeleteSong(song.GroupName, song.SongName)
		require.NoError(t, err)
	}

	// Проверяем, что библиотека пустая
	emptyLibrary, err := db.GetLibrary(ctx, domain.SongFilter{})
	require.NoError(t, err)
	require.Empty(t, emptyLibrary)

	t.Log("All tests passed successfully")
}
