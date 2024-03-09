package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/fs"
	"os"
)

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

// LoadYAMLConf YAML format is required
// conf must be pointer
func LoadYAMLConf(name string, conf interface{}) error {
	data, err := os.ReadFile(name)
	if err != nil {
		return fmt.Errorf("read yaml file=%q: %w", name, err)
	}
	err = yaml.Unmarshal(data, conf)
	if err != nil {
		return fmt.Errorf("unmarshal yaml file=%q: %w", name, err)
	}
	return nil
}

// LoadFSYAMLConf YAML format is required
// conf must be pointer
func LoadFSYAMLConf(fsys fs.FS, name string, conf interface{}) error {
	data, err := fs.ReadFile(fsys, name)
	if err != nil {
		return fmt.Errorf("read yaml file=%q: %w", name, err)
	}
	err = yaml.Unmarshal(data, conf)
	if err != nil {
		return fmt.Errorf("unmarshal YAML file=%q: %w", name, err)
	}
	return nil
}

type YamlCfg struct {
	Filename string
	FS       fs.ReadFileFS
}

func (yc YamlCfg) Load() (AppConfig, error) {
	var cfg AppConfig

	var err error
	if yc.FS == nil {
		err = LoadYAMLConf(yc.Filename, &cfg)
		return cfg, err
	}
	err = LoadFSYAMLConf(yc.FS, yc.Filename, &cfg)
	return cfg, err
}
