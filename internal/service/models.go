package service

type URL struct {
	Url string `json:"url"`
}

type Response struct {
	Result string `json:"result"`
}

type BatchUrl struct {
	Id       string `json:"correlation_id"`
	Original string `json:"original_url"`
}

type BatchResponse struct {
	Id    string `json:"correlation_id"`
	Short string `json:"short_url"`
}

type Conf struct {
	HostAddress   string
	TargetAddress string
}
