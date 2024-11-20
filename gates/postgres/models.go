package postgres

import "time"

type Song struct {
	Group       string
	SongName    string
	ReleaseDate time.Time
	Text        string
	Link        string
}
