package postgres

type Song struct {
	GroupName   string `db:"group_name"`
	SongName    string `db:"song"`
	ReleaseDate string `db:"release_date"`
	Text        string `db:"text"`
	Link        string `db:"link"`
}

type SongFilter struct {
	Group       string // Фильтр по группе
	SongName    string // Фильтр по названию песни
	ReleaseDate string // Фильтр по дате выпуска
	Limit       int    // Количество записей на странице
	Offset      int    // Сдвиг для пагинации
}
