-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS public.shorten_urls
		(
			uuid SERIAL,
			short_url text NOT NULL,
			original_url text NOT NULL
		);
-- +goose StatementEnd

-- +goose Down
