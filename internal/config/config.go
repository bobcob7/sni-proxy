package config

import (
	"encoding/json"
	"flag"
	"os"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "config.json", "Path for a JSON formatted config file")
}

func GetConfig(config interface{}) error {
	if !flag.Parsed() {
		flag.Parse()
	}
	f, err := os.Open(configPath)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(config)
}
