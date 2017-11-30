package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/pkg/errors"
)

// Configuration static vars
var (
	envFile = readEnv()

	envFileErr     error
	missingEnvVars []string
)

func readEnv() map[string]string {
	envf := os.Getenv("ENV_FILE")
	if envf == "" {
		return nil
	}
	f, err := os.Open(envf)
	if err != nil {
		envFileErr = errors.WithMessage(err, "could not open ENV_FILE")
		return nil
	}
	defer deferClose(&envFileErr, f.Close)

	m := map[string]string{}
	dec := json.NewDecoder(f)
	envFileErr = dec.Decode(&m)
	return m
}

func GetEnv(key string) string {
	s := os.Getenv(key)
	if s == "" {
		s = envFile[key]
	}
	return s
}

func MustGetEnv(key string) string {
	s := GetEnv(key)
	if s == "" {
		missingEnvVars = append(missingEnvVars, key)
	}
	return s
}

func EnvErrors() error {
	if envFileErr != nil {
		return envFileErr
	}

	if len(missingEnvVars) < 1 {
		return nil
	}

	return fmt.Errorf("missing required env var configuration: %v", missingEnvVars)
}
