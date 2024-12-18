-- +goose Up
-- Создаем таблицу songs
CREATE TABLE songs_library (
    group_name VARCHAR(255) NOT NULL,
    song VARCHAR(255) NOT NULL,
    release_date VARCHAR(255) NOT NULL,
    text TEXT NOT NULL,
    link VARCHAR(500),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (group_name, song)
);
-- +goose Down

-- Удаляем триггер и функцию
DROP TRIGGER IF EXISTS set_updated_at ON songs_library;
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Удаляем таблицу songs_library
DROP TABLE IF EXISTS songs_library;
