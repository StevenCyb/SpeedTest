package utils

import (
	"log"
	"os"
	"strconv"
)

const (
	numberBase  = 10
	bitSizeUint = 32
)

// EnvType defines supported types for `GetEnvOrDie`.
type EnvType interface {
	string | uint | bool
}

// GetEnv returns env var, default var or panics if not present but required.
func GetEnv[T EnvType](key string, defaultValue string, required bool) T { //nolint:ireturn
	var ret T

	value, exists := os.LookupEnv(key)
	if !exists {
		value = defaultValue

		if required {
			log.Panicf("missing environment variable '%s=%s'", key, value)
		}
	}

	switch ptr := any(&ret).(type) {
	case *string:
		*ptr = value

	case *uint:
		tmp, err := strconv.ParseInt(value, numberBase, bitSizeUint)
		if err != nil {
			log.Panicf("environment variable '%s=%s' is not uint", key, value)
		}

		*ptr = uint(tmp)

	case *bool:
		tmp, err := strconv.ParseBool(value)
		if err != nil {
			log.Panicf("environment variable '%s=%s' is not bool", key, value)
		}

		*ptr = tmp
	}

	return ret
}
