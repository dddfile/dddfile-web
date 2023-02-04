package util

import (
	"os"
)

func GetEnvVar(key string) string {
	return os.Getenv(key)
}

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}
