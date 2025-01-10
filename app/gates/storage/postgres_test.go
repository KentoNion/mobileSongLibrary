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

	//Считываем конфиг
	os.Setenv("CONFIG_PATH", "../../../config.yaml")
	cfg := config.MustLoad()

	//инициализируем логгер
	log := logger.MustInitLogger(cfg)

	//подключение к бд и накатывание миграции для теста
	connStr := fmt.Sprintf("user=%s password=%s dbname=mobile_song host=%s sslmode=%s timezone=UTC", cfg.DB.User, cfg.DB.Pass, cfg.DB.Host, cfg.DB.Ssl) //todo:  надо бы сюда конфиг пробросить и настроить отдельную тестовую дб в идеале
	conn, err := sqlx.Connect("postgres", connStr)                                                                                                      //подключение к бд
	if err != nil {
		panic(err)
	}
	goose.Up(conn.DB, "./migrations")

	fmt.Println("testbd migrations applied successfully")
	//подключение к бд
	require.NoError(t, err)
	db := NewDB(conn, log)

	//создание тестовых песен для загрузки в бд
	testSongs := []domain.Song{
		{
			GroupName:   "muse",
			SongName:    "Supermassive Black Hole",
			ReleaseDate: time.Date(2006, time.July, 16, 0, 0, 0, 0, time.UTC),
			Text:        "Ooh baby, don't you know I suffer?\nOoh baby, can you hear me moan?\nYou caught me under false pretenses\nHow long before you let me go?\n\nOoh\nYou set my soul alight\nOoh\nYou set my soul alight",
			Link:        "https://www.youtube.com/watch?v=Xsp3_a-PMTw",
		},
		{
			GroupName:   "muse",
			SongName:    "WON'T STAND DOWN",
			ReleaseDate: time.Date(2022, time.January, 13, 0, 0, 0, 0, time.UTC),
			Text:        "I never believed that I would concede and let someone trample on me\nYou strung me along, I thought I was strong, but you were just gaslighting me\nI've opened my eyes, and counted the lies, and now it is clearer to me\nYou are just a user, and an abuser, living vicariously\n\nWon’t stand down\nI’m growing stronger\nWon’t stand down\nI’m owned no longer\nWon’t stand down\nYou’ve used me for too long, now die alone",
			Link:        "https://youtu.be/d55ELY17CFM",
		},
		{
			GroupName:   "Buku",
			SongName:    "Front to Back",
			ReleaseDate: time.Date(2016, time.August, 30, 0, 0, 0, 0, time.UTC),
			Text:        "Front to the back, front to back\nFront to the back, front to back...",
			Link:        "https://www.youtube.com/watch?v=PWROws51oWM",
		},
	}

	//грузим тестовые песни в бд
	for _, testSong := range testSongs {
		err = db.AddSong(testSong)
		require.NoError(t, err)
	}
	//проверяем метод update
	err = db.UpdateSong(domain.Song{
		GroupName:   "Buku",
		SongName:    "Front to Back",
		ReleaseDate: time.Date(2006, time.July, 16, 0, 0, 0, 0, time.UTC),
	})
	require.NoError(t, err)
	songFromDB, err := db.GetSong("Buku", "Front to Back") //заодно проверяем метод getSong
	require.NoError(t, err)
	require.Equal(t, time.Date(2006, time.July, 16, 0, 0, 0, 0, time.UTC), songFromDB.ReleaseDate)

	//проверка переиминования группы
	err = db.GroupRename("muse", "Muse")
	require.NoError(t, err)

	//получаем список всех песен, и проверяем работу фильтра по группе (новое имя группы)
	testLibrary, err := db.GetLibrary(ctx, domain.SongFilter{GroupName: "Muse"})
	require.NoError(t, err)
	require.Equal(t, time.Date(2022, time.January, 13, 0, 0, 0, 0, time.UTC), testLibrary[1].ReleaseDate)

	//удаляем и проверяем что бд пустая
	err = db.DeleteSong("Muse", "Supermassive Black Hole")
	err = db.DeleteSong("Muse", "WON'T STAND DOWN")
	err = db.DeleteSong("Buku", "Front to Back")
	require.NoError(t, err)
	testLibrary, err = db.GetLibrary(ctx, domain.SongFilter{})
	require.NoError(t, err)
	var emptyLibrary []Song
	require.Equal(t, emptyLibrary, testLibrary)
}
