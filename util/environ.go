package util

import (
	"log"
	"os"
	"strings"
)

const sqlPW = "SQL_PW"

func Env() string {
	environ := os.Getenv("EXEC_ENVIRONMENT")
	if environ == "" {
		environ = "dev"
	}
	return environ
}

func SQL_Pw() string {
	all := os.Environ()
	found := false
	for _, v := range all {
		if strings.HasPrefix(v, sqlPW) {
			found = true
		}
	}
	if !found {
		log.Printf("\nProgram *requires* environment variable %q.\nExiting.\n", sqlPW)
		os.Exit(1)
	}

	pass := os.Getenv(sqlPW)
	return pass
}
