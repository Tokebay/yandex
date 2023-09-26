package models

type Response struct {
	Result string `json:"result"`
}

type Request struct {
	URL string `json:"url"`
}

// request
type BatchShortenRequest []struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// response
type BatchShortenResponse []struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}
