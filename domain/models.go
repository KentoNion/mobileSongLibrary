package domain

import "errors"

type Song struct {
	GroupName   string `json:"group"`
	SongName    string `json:"song"`
	ReleaseDate string `json:"release_date,omitempty"`
	Text        string `json:"text,omitempty"`
	Link        string `json:"link,omitempty"`
}

// Структура реализующая фильтры
type SongFilter struct {
	GroupName   string `db:"group_name" json:"group"`
	SongName    string `db:"song" json:"song"`
	ReleaseDate string `db:"release_date" json:"release_date,omitempty"`
	Limit       int    `json:"limit,omitempty"`
	Offset      int    `json:"offset,omitempty"`
}

func (s *Song) Validate() error {
	if s.GroupName == "" {
		return errors.New("group_name is required")
	}
	if s.SongName == "" {
		return errors.New("song_name is required")
	}
	return nil
}
