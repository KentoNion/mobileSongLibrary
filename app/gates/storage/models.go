package storage

import (
	"errors"
	"mobileSongLibrary/domain"
	"time"
)

type Song struct {
	GroupName   domain.GroupName `db:"group_name"`
	SongName    domain.SongName  `db:"song"`
	ReleaseDate time.Time        `db:"release_date"`
	Text        string           `db:"text"`
	Link        domain.Link      `db:"link"`
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

func ToStorage(dsong domain.Song) Song {
	return Song{
		GroupName:   dsong.GroupName,
		SongName:    dsong.SongName,
		ReleaseDate: time.Time(dsong.ReleaseDate),
		Text:        dsong.Text,
		Link:        dsong.Link,
	}
}

func ToDomain(ssong Song) domain.Song {
	return domain.Song{
		GroupName:   ssong.GroupName,
		SongName:    ssong.SongName,
		ReleaseDate: domain.CustomDate(ssong.ReleaseDate),
		Text:        ssong.Text,
		Link:        ssong.Link,
	}
}
