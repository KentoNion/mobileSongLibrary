package postgres

import "errors"

type Song struct {
	GroupName   string `db:"group_name" json:"group"`
	SongName    string `db:"song" json:"song"`
	ReleaseDate string `db:"release_date" json:"release_date,omitempty"`
	Text        string `db:"text" json:"text,omitempty"`
	Link        string `db:"link" json:"link,omitempty"`
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
