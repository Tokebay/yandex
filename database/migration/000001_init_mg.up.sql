CREATE TABLE IF NOT EXISTS public.shorten_urls
(
    uuid SERIAL,
    short_url text NOT NULL,
    original_url text NOT NULL
);
CREATE UNIQUE INDEX original_url_index ON shorten_urls (original_url);