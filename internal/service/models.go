package service

type URL struct {
	Url string `json:"url"`
}

type Result struct {
	Result string `json:"result"`
}

type Conf struct {
	HostAddress   string
	TargetAddress string
}
