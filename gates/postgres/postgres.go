package postgres

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/bool64/sqluct"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type DB struct {
	db *sqlx.DB
	sq sq.StatementBuilderType
	sm sqluct.Mapper
}

func NewDB(db *sqlx.DB) *DB {
	return &DB{
		db: db,
		sq: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
		sm: sqluct.Mapper{Dialect: sqluct.DialectSQLite3},
	}
}

func (p *DB) addSong(song Song) error { //функция добавления новой песни
	query := p.sm.Insert(p.sq.Insert("songs_library"), song, sqluct.InsertIgnore)
	qry, args, err := query.ToSql()
	if err != nil {
		return errors.Wrap(err, "failed to make query while adding song")
	}
	rows, err := p.db.Exec(qry, args...)
	//created_at устанавливается с помощью триггера
	if err != nil {
		return errors.Wrap(err, "failed to add song")
	}
	if rows, _ := rows.RowsAffected(); rows == 0 { //проверка на поменялось ли что
		return errors.New("failed to add song, no rows affected")
	}
	return nil
}

func (p *DB) updateSong(song Song) error {
	query := p.sq.Update("songs_library")
	if song.Link != "" { //проверка на то что линка не пустая
		query = query.Set("link", song.Link).
			Where(sq.Eq{"group": song.Group, "song": song.SongName})
	}
	if !song.ReleaseDate.IsZero() {
		query = query.Set("release_date", song.ReleaseDate).
			Where(sq.Eq{"group": song.Group, "song": song.SongName})
	}
	if song.Text != "" {
		query = query.Set("text", song.Text).
			Where(sq.Eq{"group": song.Group, "song": song.SongName})
	}
	qry, args, err := query.ToSql()
	_, err = p.db.Exec(qry, args...)
	//updated_at обновляется с помощью триггера
	if err != nil {
		return errors.Wrap(err, "failed to update song")
	}
	return nil
}

func (p *DB) GroupRename(song Song) error {}

func (p *DB) selectSong(group string, song string) error {

}

func (p *DB) deleteSong(group string, song string) error {

}
