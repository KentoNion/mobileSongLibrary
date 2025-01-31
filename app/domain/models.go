package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrCantReplaceWithEmptyRows = errors.New("Can't replace any felds with no info")

type GroupName string
type SongName string
type Link string
type CustomDate time.Time

type Song struct {
	GroupName   GroupName  `json:"group"`
	SongName    SongName   `json:"song"`
	ReleaseDate CustomDate `json:"release_date,omitempty"`
	Text        string     `json:"text,omitempty"`
	Link        Link       `json:"link,omitempty"`
}

// Структура реализующая фильтры
type SongFilter struct {
	GroupName   string     `db:"group_name" json:"group"`
	SongName    string     `db:"song" json:"song"`
	ReleaseDate CustomDate `db:"release_date" json:"release_date,omitempty"`
	Text        string     `db:"text" json:"text,omitempty"`
	Link        Link       `db:"link" json:"link,omitempty"`
	Limit       int        `json:"limit,omitempty"`
	Offset      int        `json:"offset,omitempty"`
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

func ParseCustomDate(dateStr string) (CustomDate, error) {
	const customDateFormat = "02.01.2006"
	parsedTime, err := time.Parse(customDateFormat, dateStr)
	if err != nil {
		return CustomDate{}, err
	}
	return CustomDate(parsedTime), nil
}

// Процесс маршализации и демаршализации json для даты
const customDateFormat = "02.01.2006"

func (cd CustomDate) MarshalJSON() ([]byte, error) {
	t := time.Time(cd)
	return []byte(fmt.Sprintf("\"%s\"", t.Format(customDateFormat))), nil
}

func (cd *CustomDate) UnmarshalJSON(data []byte) error {
	str := strings.Trim(string(data), "\"")
	t, err := time.Parse(customDateFormat, str)
	if err != nil {
		return err
	}
	*cd = CustomDate(t)
	return nil
}
