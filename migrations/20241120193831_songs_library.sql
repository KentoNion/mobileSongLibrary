-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE songs (
    id SERIAL PRIMARY KEY,
    group VARCHAR(255) NOT NULL,
    song VARCHAR(255) NOT NULL,
    release_date DATE NOT NULL,
    text TEXT NOT NULL,
    link VARCHAR(500),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE songs_library;
-- +goose StatementEnd
