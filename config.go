package main

import (
	"fmt"
	"os"
)

var missingEnvVars []string

func GetEnv(key string) string {
	s := os.Getenv(key)
	if s == "" {
		missingEnvVars = append(missingEnvVars, key)
	}
	return s
}

func EnvErrors() error {
	if len(missingEnvVars) < 1 {
		return nil
	}

	return fmt.Errorf("missing required env var configuration: %v", missingEnvVars)
}
