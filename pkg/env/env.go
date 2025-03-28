package env

import (
	"fmt"
	"os"
)

func Get(name string) (string, error) {
	value := os.Getenv(name)
	if value == "" {
		return "", fmt.Errorf("missing or empty environment variable: '%s' is required", name)
	}

	return value, nil
}
