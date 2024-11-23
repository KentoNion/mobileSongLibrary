package postgres

import "time"

type Song struct {
	Group       string
	SongName    string
	ReleaseDate time.Time
	Text        string
	Link        string
}

type SongFilter struct {
	Group       *string    // Фильтр по группе
	SongName    *string    // Фильтр по названию песни
	ReleaseDate *time.Time // Фильтр по дате выпуска
	Limit       int        // Количество записей на странице
	Offset      int        // Сдвиг для пагинации
}
