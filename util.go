package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

func GetLogFile(name string) (*os.File, error) {
	file, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	return file, nil
}

// LoadYAMLConf YAML format is required
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
