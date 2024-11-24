package postgres

import (
	"context"
	sq "github.com/Masterminds/squirrel"
	"github.com/bool64/sqluct"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"time"
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
		sm: sqluct.Mapper{Dialect: sqluct.DialectPostgres},
	}
}

func (p *DB) AddSong(song Song) error { //функция добавления новой песни
	query := p.sq.Insert("songs_library").
		Columns("group_name", "song", "release_date", "text", "link", "created_at", "updated_at").
		Values(song.GroupName, song.SongName, song.ReleaseDate, song.Text, song.Link, time.Now(), time.Now()).
		Suffix("ON CONFLICT (group_name, song) DO NOTHING")
	qry, args, err := query.ToSql()
	if err != nil {
		return errors.Wrap(err, "failed to make query while adding song")
	}
	rows, err := p.db.Exec(qry, args...)

	if err != nil {
		return errors.Wrap(err, "failed to add song")
	}
	if affectedRows, _ := rows.RowsAffected(); affectedRows == 0 { //проверка на поменялось ли что
		return errors.New("failed to add song, no rows affected")
	}
	return nil
}

func (p *DB) UpdateSong(song Song) error {
	query := p.sq.Update("songs_library")
	if song.Link != "" { //проверка на то что линка не пустая
		query = query.Set("link", song.Link).
			Where(sq.Eq{"group_name": song.GroupName, "song": song.SongName})
	}
	if song.ReleaseDate != "" {
		query = query.Set("release_date", song.ReleaseDate).
			Where(sq.Eq{"group_name": song.GroupName, "song": song.SongName})
	}
	if song.Text != "" {
		query = query.Set("text", song.Text).
			Where(sq.Eq{"group_name": song.GroupName, "song": song.SongName})
	}
	query = query.Set("updated_at", time.Now()).
		Where(sq.Eq{"group_name": song.GroupName, "song": song.SongName})
	qry, args, err := query.ToSql()
	_, err = p.db.Exec(qry, args...)
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
	err = p.db.Get(&result, qry, args...)
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

func (p *DB) GetLibrary(ctx context.Context, filter SongFilter) ([]Song, error) {
	// Создаем базовый запрос
	query := p.sm.Select(p.sq.Select(), &Song{}).From("songs_library")

	// Фильтрация по значениям фильтра
	if filter.GroupName != "" {
		query = query.Where("group_name = ?", filter.GroupName)
	}
	if filter.SongName != "" {
		query = query.Where("song = ?", filter.SongName)
	}
	if filter.ReleaseDate != "" {
		query = query.Where("release_date = ?", filter.ReleaseDate)
	}

	// Пагинация (если задан лимит)
	if filter.Limit > 0 {
		query = query.Limit(uint64(filter.Limit)).Offset(uint64(filter.Offset))
	}

	// Генерация SQL-запроса
	qry, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build query for songs_library")
	}

	// Выполняем запрос
	var songs []Song
	err = p.db.SelectContext(ctx, &songs, qry, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch songs from songs_library")
	}

	return songs, nil
}
