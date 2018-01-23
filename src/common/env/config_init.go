package env

import (
	"github.com/naoina/toml"
	"os"
)

func InitConfig(ConfPath string, config interface{}) error {

	f, err := os.Open(ConfPath)

	if err != nil {
		return err
	}
	defer f.Close()

	if err := toml.NewDecoder(f).Decode(config); err != nil {
		return err
	}

	return nil
}
