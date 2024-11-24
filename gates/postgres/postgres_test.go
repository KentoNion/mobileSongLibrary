package postgres

import (
	"context"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" //драйвер postgres
	"github.com/stretchr/testify/require"
	"testing"
)

func TestInsertUpdateSelectGetLibraryRenameGroupDelete(t *testing.T) {
	ctx := context.Background()
	//подключение к бд
	conn, err := sqlx.Connect("postgres", "user=postgres password=postgres dbname=testdb host=localhost sslmode=disable")
	require.NoError(t, err)
	db := NewDB(conn)

	//создание тестовых песен для загрузки в бд
	testSongs := []Song{
		{
			GroupName:   "muse",
			SongName:    "Supermassive Black Hole",
			ReleaseDate: "16.07.2006",
			Text:        "Ooh baby, don't you know I suffer?\nOoh baby, can you hear me moan?\nYou caught me under false pretenses\nHow long before you let me go?\n\nOoh\nYou set my soul alight\nOoh\nYou set my soul alight",
			Link:        "https://www.youtube.com/watch?v=Xsp3_a-PMTw",
		},
		{
			GroupName:   "muse",
			SongName:    "WON'T STAND DOWN",
			ReleaseDate: "13.01.2022",
			Text:        "I never believed that I would concede and let someone trample on me\nYou strung me along, I thought I was strong, but you were just gaslighting me\nI've opened my eyes, and counted the lies, and now it is clearer to me\nYou are just a user, and an abuser, living vicariously\n\nWon’t stand down\nI’m growing stronger\nWon’t stand down\nI’m owned no longer\nWon’t stand down\nYou’ve used me for too long, now die alone",
			Link:        "https://youtu.be/d55ELY17CFM",
		},
		{
			GroupName:   "Buku",
			SongName:    "Front to Back",
			ReleaseDate: "30.08.2016",
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
	err = db.UpdateSong(Song{
		GroupName:   "Buku",
		SongName:    "Front to Back",
		ReleaseDate: "31.08.2016",
	})
	require.NoError(t, err)
	songFromDB, err := db.GetSong("Buku", "Front to Back") //заодно проверяем метод getSong
	require.NoError(t, err)
	require.Equal(t, "31.08.2016", songFromDB.ReleaseDate)

	//проверка переиминования группы
	err = db.GroupRename("muse", "Muse")
	require.NoError(t, err)

	//получаем список всех песен, и проверяем работу фильтра по группе (новое имя группы)
	testLibrary, err := db.GetLibrary(ctx, SongFilter{GroupName: "Muse"})
	require.NoError(t, err)
	require.Equal(t, "13.01.2022", testLibrary[1].ReleaseDate)

	//удаляем и проверяем что бд пустая
	err = db.DeleteSong("Muse", "Supermassive Black Hole")
	err = db.DeleteSong("Muse", "WON'T STAND DOWN")
	err = db.DeleteSong("Buku", "Front to Back")
	require.NoError(t, err)
	testLibrary, err = db.GetLibrary(ctx, SongFilter{})
	require.NoError(t, err)
	var emptyLibrary []Song
	require.Equal(t, emptyLibrary, testLibrary)
}
