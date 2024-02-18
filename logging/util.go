package logging

import (
	"fmt"
	"os"
)

func Must(file *os.File, err error) *os.File {
	if err != nil {
		panic(err)
	}
	return file
}

func GetLogFile(name string) (*os.File, error) {
	file, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	return file, nil
}
