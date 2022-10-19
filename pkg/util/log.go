package util

import (
	"os"

	"github.com/rs/zerolog/log"
)

// Check env variable
func IsDebugging() bool {
	return os.Getenv("GALAX_DEBUGGING") == "true"
}

func SetDebugging(debugging bool) {
	if debugging {
		os.Setenv("GALAX_DEBUGGING", "true")
	} else {
		os.Setenv("GALAX_DEBUGGING", "false")
	}
}

func DebugOutput(format string, a ...interface{}) {
	if IsDebugging() {
		log.Debug().Msgf(format, a...)
	}
}
