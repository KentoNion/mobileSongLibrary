package postgres

type Song struct {
	GroupName   string `db:"group_name" json:"group"`
	SongName    string `db:"song" json:"song"`
	ReleaseDate string `db:"release_date" json:"release_date,omitempty"`
	Text        string `db:"text" json:"text,omitempty"`
	Link        string `db:"link" json:"link,omitempty"`
}

type SongFilter struct {
	GroupName   string `db:"group_name" json:"group"`                    // Фильтр по группе
	SongName    string `db:"song" json:"song"`                           // Фильтр по названию песни
	ReleaseDate string `db:"release_date" json:"release_date,omitempty"` // Фильтр по дате выпуска
	Limit       int    `json:"limit,omitempty"`                          // Количество записей на странице
	Offset      int    `json:"offset,omitempty"`                         // Сдвиг для пагинации
}
