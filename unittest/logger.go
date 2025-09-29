package unittest

import (
	"github.com/rs/zerolog"
	"os"
)

// Logger returns a zerolog.Logger that writes to stdout with the given level and timestamps.
func Logger(level zerolog.Level) zerolog.Logger {
	return zerolog.New(os.Stdout).Level(level).With().Timestamp().Logger()
}
