package config_impl

import (
	"Yandex/internal/conf"
	"flag"
	"net/url"
	"os"
)

const defaultAddress = "localhost:8888"

var _ conf.Config = (*ConfigImpl)(nil)

type ConfigImpl struct {
	hostAddress   string
	targetAddress string
}

func (c *ConfigImpl) GetHostAddress() string {
	return c.hostAddress
}

func (c *ConfigImpl) GetTargetAddress() string {
	return c.targetAddress
}

func New() *ConfigImpl {
	ha := getArg("a", "SERVER_ADDRESS", "Address where to start http server")
	ht := getArg("b", "BASE_URL", "Address to send short urls")
	flag.Parse()
	return &ConfigImpl{
		hostAddress:   *ha,
		targetAddress: *ht,
	}
}

func getArg(flagName, env, usage string) (res *string) {
	address := os.Getenv(env)
	res = &address
	tmp := flag.String(flagName, defaultAddress, usage)
	if address == "" || !checkIfAddress(address) {
		res = tmp
	}
	return
}

func checkIfAddress(address string) bool {
	_, err := url.Parse(address)
	return err == nil
}
