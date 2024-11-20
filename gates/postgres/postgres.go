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
	if err != nil {
		return errors.Wrap(err, "failed to add song")
	}
	if rows, _ := rows.RowsAffected(); rows == 0 { //проверка на поменялось ли что
		return errors.New("failed to add song, no rows affected")
	}
	return nil
}

func (p *DB) updateSong(song Song) error {
	query := `
	UPDATE songs_library
	SET 
		release_date = $3,
		text = $4
		link = $5,
		updated_at = NOW()
	WHERE "group" = $1 AND song = $2;
`
	_, err := p.db.Exec(query, song.Group, song.SongName, song.ReleaseDate, song.Text, song.Link)
	//todo а если какой то из параметров окажется пустым? Надо не перезаписывать его
	if err != nil {
		return errors.Wrap(err, "failed to update song")
	}
	return nil
}

func (p *DB) selectSong(group string, song string) error {

}

func (p *DB) deleteSong(group string, song string) error {

}
