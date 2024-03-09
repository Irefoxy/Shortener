package conf

import (
	"Yandex/internal/service"
	"flag"
	"os"
)

const (
	defaultAddress = "localhost:8888"
	defaultPath    = "/tmp/short-url-db.json"
)

type ConfigImpl struct {
	service      service.Conf
	fileLocation string
}

func (c *ConfigImpl) GetServiceConf() *service.Conf {
	return &c.service
}

func (c *ConfigImpl) GetFileLocation() string {
	return c.fileLocation
}

func New() *ConfigImpl {
	ha := getArg("a", "SERVER_ADDRESS", "Address where to start http server", defaultAddress)
	ta := getArg("b", "BASE_URL", "Address to send short urls", defaultAddress)
	fl := getArg("f", "FILE_STORAGE_PATH", "Location of storage file", defaultPath)
	flag.Parse()
	return &ConfigImpl{
		service: service.Conf{
			HostAddress:   *ha,
			TargetAddress: *ta,
		},
		fileLocation: *fl,
	}
}

func getArg(flagName, env, usage, def string) (res *string) {
	address := os.Getenv(env)
	res = &address
	tmp := flag.String(flagName, def, usage)
	if address == "" {
		res = tmp
	}
	return
}
