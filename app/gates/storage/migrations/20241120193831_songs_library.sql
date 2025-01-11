-- +goose Up
ALTER DATABASE songs_library SET timezone TO 'UTC';
-- Создаем таблицу songs
CREATE TABLE songs_library (
    group_name VARCHAR(255) NOT NULL,
    song VARCHAR(255) NOT NULL,
    release_date TIMESTAMP WITH TIME ZONE,
    text TEXT,
    link VARCHAR(500),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (group_name, song)
);
--выставил индексы на самые частые варианты поиска песни
CREATE INDEX idx_song ON songs_library(song);
CREATE INDEX idx_group ON songs_library(group_name);
-- +goose Down
-- Удаляем таблицу songs_library
DROP INDEX IF EXISTS idx_song;
DROP INDEX IF EXISTS idx_group;
DROP TABLE IF EXISTS songs_library;
