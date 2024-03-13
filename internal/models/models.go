package models

type Plain string

func (p Plain) String() string {
	return string(p)
}

type URL struct {
	Url string `json:"url"`
}

func (u URL) String() string {
	return u.Url
}

type ShortURL struct {
	Result string `json:"result"`
}

type BatchURL struct {
	Id       string `json:"correlation_id"`
	Original string `json:"original_url"`
}

func (u BatchURL) String() string {
	return u.Original
}

type BatchShortURL struct {
	Id    string `json:"correlation_id"`
	Short string `json:"short_url"`
}

type ServiceUnit struct {
	Id          string
	OriginalUrl string
	ShortUrl    string
}

type Conf struct {
	HostAddress   string
	TargetAddress string
}
