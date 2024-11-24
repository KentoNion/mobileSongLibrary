package postgres

type Song struct {
	GroupName   string `db:"group_name"`
	SongName    string `db:"song"`
	ReleaseDate string `db:"release_date"`
	Text        string `db:"text"`
	Link        string `db:"link"`
}

type SongFilter struct {
	GroupName   string `db:"group_name"`   // Фильтр по группе
	SongName    string `db:"song"`         // Фильтр по названию песни
	ReleaseDate string `db:"release_date"` // Фильтр по дате выпуска
	Limit       int    // Количество записей на странице
	Offset      int    // Сдвиг для пагинации
}
