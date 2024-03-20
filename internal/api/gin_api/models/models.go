package models

type URL struct {
	Url string `json:"url"`
}

type ShortURL struct {
	Result string `json:"result"`
}

type BatchURL struct {
	Id       string `json:"correlation_id"`
	Original string `json:"original_url"`
}

type BatchShortURL struct {
	Id    string `json:"correlation_id"`
	Short string `json:"short_url"`
}
