package tigerblood

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

type Config struct {
	Credentials   map[string]string
	BindAddress   string
	DatabaseDsn   string
	StatsdAddress string
	EnableHawk    bool
}

// LoadConfigFromPath loads a yaml config file from the provided path, overriding values based on environment variables
func LoadConfigFromPath(path string) (Config, error) {
	var config Config
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return config, err
	}
	err = yaml.Unmarshal(bytes, &config)
	if dsn, found := os.LookupEnv("TIGERBLOOD_DSN"); found {
		config.DatabaseDsn = dsn
	}
	if statsdAddr, found := os.LookupEnv("TIGERBLOOD_STATSD_ADDR"); found {
		config.StatsdAddress = statsdAddr
	}
	if bindAddr, found := os.LookupEnv("TIGERBLOOD_BIND_ADDR"); found {
		config.BindAddress = bindAddr
	}
	if hawk, _ := os.LookupEnv("TIGERBLOOD_ENABLE_HAWK"); hawk == "yes" {
		config.EnableHawk = true
	}
	return config, err
}
