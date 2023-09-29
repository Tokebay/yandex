-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS public.shorten_urls
(
	uuid SERIAL,
	short_url text NOT NULL,
	original_url text NOT NULL,
	userID int,
	FOREIGN KEY (user_id) REFERENCES users (user_id)
);

CREATE TABLE users
(
    user_id serial PRIMARY key
);
-- +goose StatementEnd

-- +goose Down
