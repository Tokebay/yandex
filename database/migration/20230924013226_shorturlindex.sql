-- +goose Up
-- +goose StatementBegin
CREATE UNIQUE INDEX original_url_index ON shorten_urls (original_url);
-- +goose StatementEnd

-- +goose Down

