package util

import (
	"log"
	"os"
	"strings"
)

func Env() string {
	environ := os.Getenv("EXEC_ENVIRONMENT")
	if environ == "" {
		environ = "dev"
	}
	return environ
}

func EnvVar(key string) string {
	all := os.Environ()
	found := false
	for _, v := range all {
		if strings.HasPrefix(v, key) {
			found = true
		}
	}
	if !found {
		log.Printf("\nProgram *requires* environment variable %q.\nExiting.\n", key)
		os.Exit(1)
	}

	envVal := os.Getenv(key)
	return envVal
}
