package env

import (
	"log"
	"os"
	"strconv"
)

func GetString(key, fallback string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	return value
}

func GetInt(key string, fallback int) int {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	valueAsInt, err := strconv.Atoi(value)
	if err != nil {
		log.Println(err)
		return fallback
	}
	return valueAsInt
}
