package models

type ShortenURL struct {
	UUID        int
	ShortURL    string
	OriginalURL string
	UserID      int
	DeletedFlag bool
}
