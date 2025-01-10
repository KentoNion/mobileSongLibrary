package domain

import (
	"errors"
	"time"
)

type GroupName string
type SongName string
type Link string

type Song struct {
	GroupName   GroupName `json:"group"`
	SongName    SongName  `json:"song"`
	ReleaseDate time.Time `json:"release_date,omitempty"`
	Text        string    `json:"text,omitempty"`
	Link        Link      `json:"link,omitempty"`
}

// Структура реализующая фильтры
type SongFilter struct {
	GroupName   string    `db:"group_name" json:"group"`
	SongName    string    `db:"song" json:"song"`
	ReleaseDate time.Time `db:"release_date" json:"release_date,omitempty"`
	Limit       int       `json:"limit,omitempty"`
	Offset      int       `json:"offset,omitempty"`
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

func ParseTime(timeStr string) (time.Time, error) {
	parsedTime, err := time.Parse("02.01.2006", timeStr)
	if err != nil {
		return time.Time{}, err
	}
	return parsedTime, nil
}
