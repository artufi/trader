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
		return fmt.Errorf("failed to read file=%q: %w", name, err)
	}
	err = yaml.Unmarshal(data, conf)
	if err != nil {
		return fmt.Errorf("failed to unmarshal YAML file=%q: %w", name, err)
	}
	return nil
}

func LoadFSYAMLConf(fsys fs.FS, name string, conf interface{}) error {
	data, err := fs.ReadFile(fsys, name)
	if err != nil {
		return fmt.Errorf("failed to read file=%q: %w", name, err)
	}
	err = yaml.Unmarshal(data, conf)
	if err != nil {
		return fmt.Errorf("failed to unmarshal YAML file=%q: %w", name, err)
	}
	return nil
}
