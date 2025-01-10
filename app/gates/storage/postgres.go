package storage

import (
	"context"
	sq "github.com/Masterminds/squirrel"
	"github.com/bool64/sqluct"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"log/slog"
	"mobileSongLibrary/domain"
	"time"
)

type DB struct {
	db  *sqlx.DB
	sq  sq.StatementBuilderType
	sm  sqluct.Mapper
	log *slog.Logger
}

func NewDB(db *sqlx.DB, log *slog.Logger) *DB {
	return &DB{
		db:  db,
		sq:  sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
		sm:  sqluct.Mapper{Dialect: sqluct.DialectPostgres},
		log: log,
	}
}

func (p *DB) AddSong(song domain.Song) error { //функция добавления новой песни
	const op = "storage.postgres.AddSong"

	p.log.Debug(op, "trying to add Song: ", song.SongName)
	query := p.sq.Insert("songs_library").
		Columns("group_name", "Song", "release_date", "text", "link", "created_at", "updated_at").
		Values(song.GroupName, song.SongName, song.ReleaseDate, song.Text, song.Link, time.Now(), time.Now()).
		Suffix("ON CONFLICT (group_name, Song) DO NOTHING")
	qry, args, err := query.ToSql()
	if err != nil {
		p.log.Error(op, " ERROR: ", err)
		return err
	}

	p.log.Debug(op, "qry: ", qry, "args: ", args)

	if err != nil {
		return errors.Wrap(err, "failed to make query while adding Song")
	}
	rows, err := p.db.Exec(qry, args...)
	if err != nil {
		p.log.Error(op, " ERROR: ", err)
		return err
	}

	if err != nil {
		return errors.Wrap(err, "failed to add Song")
	}
	if affectedRows, _ := rows.RowsAffected(); affectedRows == 0 { //проверка на поменялось ли что
		return errors.New("failed to add Song, no rows affected")
	}
	p.log.Debug(op, "Successfully added Song: ", song.SongName)
	return nil
}

func (p *DB) UpdateSong(song domain.Song) error {
	const op = "storage.postgres.UpdateSong"

	p.log.Debug(op, "trying to update Song: ", song.SongName)
	query := p.sq.Update("songs_library")

	/* -------------------------------------------------------------------------------------------------------------------
	Важное замечание!
	Я считаю что если у песни уже есть какая-либо заполненная информация, то её нужно заменять только на другую информацию
	никак не на пустое поле, чтоб нельзя было случайно удалить уже существующие данные, при редактировании других
	*/

	if song.Link != "" { //проверка на то что линка не пустая
		p.log.Debug(op, "Song link not empty, replacing with: ", song.Link)
		query = query.Set("link", song.Link).
			Where(sq.Eq{"group_name": song.GroupName, "Song": song.SongName})
	}
	if !song.ReleaseDate.IsZero() {
		p.log.Debug(op, "Song release_date not empty, replacing with: ", song.ReleaseDate)
		query = query.Set("release_date", song.ReleaseDate).
			Where(sq.Eq{"group_name": song.GroupName, "Song": song.SongName})
	}
	if song.Text != "" {
		p.log.Debug(op, "Song text not empty, replacing with: ", song.Text)
		query = query.Set("text", song.Text).
			Where(sq.Eq{"group_name": song.GroupName, "Song": song.SongName})
	}
	query = query.Set("updated_at", time.Now()).
		Where(sq.Eq{"group_name": song.GroupName, "Song": song.SongName})
	qry, args, err := query.ToSql()
	if err != nil {
		p.log.Error(op, " ERROR: ", err)
		return err
	}
	p.log.Debug(op, "qry: ", qry, "args: ", args)
	_, err = p.db.Exec(qry, args...)
	if err != nil {
		p.log.Error(op, " ERROR: ", err)
		return err
	}
	p.log.Debug(op, "Successfully updated Song: ", song.SongName)
	return nil
}

func (p *DB) GroupRename(oldGroupName string, newGroupName string) error {
	const op = "storage.postgres.GroupRename"

	p.log.Debug(op, "trying to rename group: ", oldGroupName, " to ", newGroupName)
	query := p.sq.Update("songs_library")
	query = query.Set("group_name", newGroupName).
		Where(sq.Eq{"group_name": oldGroupName})
	qry, args, err := query.ToSql()
	if err != nil {
		p.log.Error(op, " ERROR: ", err)
		return err
	}
	_, err = p.db.Exec(qry, args...)
	//updated_at обновляется с помощью триггера
	if err != nil {
		p.log.Error(op, " ERROR: ", err)
		return err
	}
	p.log.Debug(op, "Successfully renamed group: ", newGroupName)
	return nil
}

func (p *DB) GetSong(group domain.GroupName, song domain.SongName) (Song, error) {
	const op = "storage.postgres.GetSong"

	p.log.Debug(op, "trying to get Song: ", song)
	var result Song
	query := p.sm.Select(p.sq.Select(), &Song{}).
		From("songs_library").
		Where(sq.Eq{"group_name": group, "Song": song})
	qry, args, err := query.ToSql()
	if err != nil {
		p.log.Error(op, " ERROR: ", err)
		return result, err
	}
	err = p.db.Get(&result, qry, args...)
	if err != nil {
		p.log.Error(op, " ERROR: ", err)
		return result, err
	}
	result.ReleaseDate = result.ReleaseDate
	p.log.Debug(op, "Successfully retrieved Song: ", song)
	return result, nil
}

func (p *DB) DeleteSong(group domain.GroupName, song domain.SongName) error {
	const op = "storage.postgres.DeleteSong"

	p.log.Debug(op, "trying to delete Song: ", song)
	query := p.sq.Delete("songs_library").
		Where(sq.Eq{"group_name": group, "Song": song})
	qry, args, err := query.ToSql()
	if err != nil {
		p.log.Error(op, " ERROR: ", err)
		return err
	}
	p.log.Debug(op, "qry: ", qry, "args: ", args)
	_, err = p.db.Exec(qry, args...)
	if err != nil {
		p.log.Error(op, " ERROR: ", err)
		return err
	}
	p.log.Debug(op, "Successfully deleted Song: ", song)
	return nil
}

func (p *DB) GetLibrary(ctx context.Context, filter domain.SongFilter) ([]Song, error) {
	const op = "storage.postgres.GetLibrary"

	p.log.Debug(op, "trying to get songs, filter is: ", filter)
	// Создаем базовый запрос
	query := p.sm.Select(p.sq.Select(), &Song{}).From("songs_library")

	// Фильтрация по значениям фильтра
	if filter.GroupName != "" {
		query = query.Where("group_name = ?", filter.GroupName)
	}
	if filter.SongName != "" {
		query = query.Where("Song = ?", filter.SongName)
	}
	if !filter.ReleaseDate.IsZero() {
		query = query.Where("release_date = ?", filter.ReleaseDate)
	}

	// Пагинация (если задан лимит)
	if filter.Limit > 0 {
		query = query.Limit(uint64(filter.Limit)).Offset(uint64(filter.Offset))
	}

	// Генерация SQL-запроса
	qry, args, err := query.ToSql()
	if err != nil {
		p.log.Error(op, " ERROR: ", err)
		return nil, err
	}
	p.log.Debug(op, "qry: ", qry, "args: ", args)

	// Выполняем запрос
	var songs []Song
	err = p.db.SelectContext(ctx, &songs, qry, args...)
	if err != nil {
		p.log.Error(op, " ERROR: ", err)
		return nil, err
	}
	p.log.Debug(op, "Successfully retrieved songs", "")
	return songs, nil
}
