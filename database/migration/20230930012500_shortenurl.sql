-- +goose Up
-- +goose StatementBegin
CREATE TABLE users
(
    user_id serial PRIMARY key
);

CREATE TABLE IF NOT EXISTS shorten_urls
(
	uuid SERIAL,
	short_url text NOT NULL,
	original_url text NOT NULL,
	user_id int,
	FOREIGN KEY (user_id) REFERENCES users (user_id)
);

-- Удаляем существующий индекс, если он существует
DROP INDEX IF EXISTS original_url_index;

-- Создаем новый уникальный индекс
CREATE UNIQUE INDEX original_url_index ON shorten_urls (original_url);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Откатываем изменения в обратном порядке, начиная с удаления нового индекса
DROP INDEX IF EXISTS original_url_index;

-- Откатываем создание таблиц
DROP TABLE IF EXISTS shorten_urls;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
