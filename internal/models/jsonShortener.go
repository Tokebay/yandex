package models

type Response struct {
	Result string `json:"result"`
}

type Request struct {
	URL string `json:"url"`
}
