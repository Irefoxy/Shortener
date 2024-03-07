package conf

type Config interface {
	GetHostAddress() string
	GetTargetAddress() string
}
