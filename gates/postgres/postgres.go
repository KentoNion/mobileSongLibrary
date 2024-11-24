package postgres

import (
	"context"
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

func (p *DB) AddSong(song Song) error { //функция добавления новой песни
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

func (p *DB) UpdateSong(song Song) error {
	query := p.sq.Update("songs_library")
	if song.Link != "" { //проверка на то что линка не пустая
		query = query.Set("link", song.Link).
			Where(sq.Eq{"group_name": song.Group, "song": song.SongName})
	}
	if song.ReleaseDate != "" {
		query = query.Set("release_date", song.ReleaseDate).
			Where(sq.Eq{"group_name": song.Group, "song": song.SongName})
	}
	if song.Text != "" {
		query = query.Set("text", song.Text).
			Where(sq.Eq{"group_name": song.Group, "song": song.SongName})
	}
	qry, args, err := query.ToSql()
	_, err = p.db.Exec(qry, args...)
	//updated_at обновляется с помощью триггера
	if err != nil {
		return errors.Wrap(err, "failed to update song")
	}
	return nil
}

func (p *DB) GroupRename(oldGroupName string, newGroupName string) error {
	query := p.sq.Update("songs_library")
	query = query.Set("group_name", newGroupName).
		Where(sq.Eq{"group_name": oldGroupName})
	qry, args, err := query.ToSql()
	_, err = p.db.Exec(qry, args...)
	//updated_at обновляется с помощью триггера
	if err != nil {
		return errors.Wrap(err, "failed to update song")
	}
	return nil
}

func (p *DB) GetSong(group string, song string) (Song, error) {
	var result Song
	query := p.sm.Select(p.sq.Select(), &Song{}).
		From("songs_library").
		Where(sq.Eq{"group_name": group, "song": song})
	qry, args, err := query.ToSql()
	if err != nil {
		return result, errors.Wrap(err, "failed to build query to songs_library")
	}
	err = p.db.Select(&Song{}, qry, args...)
	if err != nil {
		return result, errors.Wrap(err, "failed to get song")
	}

	return result, nil
}

func (p *DB) DeleteSong(group string, song string) error {
	query := p.sq.Delete("songs_library").
		Where(sq.Eq{"group_name": group, "song": song})
	qry, args, err := query.ToSql()
	_, err = p.db.Exec(qry, args...)
	if err != nil {
		return errors.Wrap(err, "failed to delete song")
	}
	return nil
}

func (p *DB) GetLibrary(ctx context.Context, filter SongFilter) ([]Song, error) { //вывод всей библиотеки
	query := p.sm.Select(p.sq.Select(), &Song{}).From("songs_library")

	// Фильтрация
	if filter.Group != "" {
		query = query.Where("group = ?", filter.Group)
	}
	if filter.SongName != "" {
		query = query.Where("song = ?", filter.SongName)
	}
	if filter.ReleaseDate != "" {
		query = query.Where("release_date = ?", filter.ReleaseDate)
	}

	//Пагинация
	query = query.Limit(uint64(filter.Limit)).Offset(uint64(filter.Offset))

	//sql запрос
	qry, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build query to songs_library")
	}
	var songs []Song
	err = p.db.SelectContext(ctx, &songs, qry, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get filtered songs")
	}

	return songs, nil
}
