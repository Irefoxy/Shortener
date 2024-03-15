package conf

import (
	"Yandex/internal/models"
	"flag"
	"fmt"
	"os"
)

const (
	defaultAddress = "localhost:8888"
)

type ConfigImpl struct {
	service        models.Conf
	fileLocation   *string
	databaseString *string
}

func (c *ConfigImpl) GetServiceConf() *models.Conf {
	return &c.service
}

func (c *ConfigImpl) GetFileLocation() string {
	return *c.fileLocation
}

func (c *ConfigImpl) GetDatabaseString() string {
	return *c.databaseString
}

func New() *ConfigImpl {
	return &ConfigImpl{}
}

func (c *ConfigImpl) Parse(programName string, argv []string) {
	flagSet := flag.NewFlagSet(programName, flag.ContinueOnError)
	c.service.HostAddress = getArg(flagSet, "SERVER_ADDRESS", "Address where to start http server", defaultAddress, "a")
	c.service.TargetAddress = getArg(flagSet, "BASE_URL", "Address to send short urls", defaultAddress, "b")
	c.fileLocation = getArg(flagSet, "FILE_STORAGE_PATH", "Location of storage file", "", "f")
	c.databaseString = getArg(flagSet, "DATABASE_DSN", "Database config string", "", "d")
	err := flagSet.Parse(argv)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func getArg(flagSet *flag.FlagSet, env, usage, def, flagName string) *string {
	address := os.Getenv(env)
	tmp := flagSet.String(flagName, def, usage)
	if address == "" {
		return tmp
	}
	return &address
}
