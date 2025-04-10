-- +goose Up
-- +goose StatementBegin
BEGIN;

CREATE TABLE IF NOT EXISTS people (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    surname VARCHAR(255) NOT NULL,
    patronymic VARCHAR(255) NULL,
    age INTEGER NULL,
    gender VARCHAR(50) NULL,
    nationality VARCHAR(10) NULL,
    created_at TIMESTAMP ,
    updated_at TIMESTAMP 
);

-- Создание индексов
CREATE INDEX IF NOT EXISTS idx_people_name ON people (name);
CREATE INDEX IF NOT EXISTS idx_people_surname ON people (surname);
COMMIT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS people;;
-- +goose StatementEnd