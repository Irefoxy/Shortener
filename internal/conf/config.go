package conf

import (
	"Yandex/internal/models"
	"flag"
	"os"
)

const (
	defaultAddress = "localhost:8888"
)

type ConfigImpl struct {
	service        models.Conf
	fileLocation   string
	databaseString string
}

func (c *ConfigImpl) GetServiceConf() *models.Conf {
	return &c.service
}

func (c *ConfigImpl) GetFileLocation() string {
	return c.fileLocation
}

func (c *ConfigImpl) GetDatabaseString() string {
	return c.databaseString
}

func New() *ConfigImpl {
	ha := getArg("a", "SERVER_ADDRESS", "Address where to start http server", defaultAddress)
	ta := getArg("b", "BASE_URL", "Address to send short urls", defaultAddress)
	fl := getArg("f", "FILE_STORAGE_PATH", "Location of storage file", "")
	ds := getArg("d", "DATABASE_DSN", "Database config string", "")
	flag.Parse()
	return &ConfigImpl{
		service: models.Conf{
			HostAddress:   *ha,
			TargetAddress: *ta,
		},
		fileLocation:   *fl,
		databaseString: *ds,
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
