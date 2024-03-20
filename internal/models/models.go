package models

type Entry struct {
	Id          string
	OriginalUrl string
	ShortUrl    string
	DeletedFlag bool
}

type ApiConf struct {
	HostAddress   *string
	TargetAddress *string
}
