package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"
	"time"

	"github.com/natefinch/lumberjack"
	"github.com/rs/zerolog"
)

type contextKey string

const TraceIdKey contextKey = "trace_id"

var Log *zerolog.Logger

type LoggerConfig struct {
	Level      zerolog.Level
	Filename   string
	MaxSize    int
	MaxBackups int
	MaxAge     int
	Compress   bool
	IsDev      string
}

func InitLogger(cfg LoggerConfig) {
	l := NewLogger(cfg).With().Str("module", "application").Logger()
	Log = &l
}

func NewLogger(config LoggerConfig) *zerolog.Logger {
	zerolog.TimeFieldFormat = time.RFC3339

	// lvl, err := zerolog.ParseLevel(config.Level)
	// if err != nil {
	// 	lvl = zerolog.InfoLevel
	// }
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	var writer io.Writer

	if config.IsDev == "development" {
		if strings.Contains(config.Filename, "app.log") {
			writer = zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
		} else {
			writer = PrettyJSONWriter{Writer: os.Stdout}
		}
	} else {
		writer = &lumberjack.Logger{
			Filename:   config.Filename,
			MaxSize:    config.MaxSize,
			MaxBackups: config.MaxBackups,
			MaxAge:     config.MaxAge,
			Compress:   config.Compress,
		}
	}

	logger := zerolog.New(writer).With().Timestamp().Str("app", "defects-rate-monitor").Logger()

	return &logger
}

type PrettyJSONWriter struct {
	Writer io.Writer
}

func (pw PrettyJSONWriter) Write(p []byte) (n int, err error) {
	var prettyJSON bytes.Buffer

	if err := json.Indent(&prettyJSON, p, "", "  "); err != nil {
		return pw.Writer.Write(p)
	}

	return pw.Writer.Write(prettyJSON.Bytes())

}

func GetTraceID(ctx context.Context) string {
	if traceID, ok := ctx.Value(TraceIdKey).(string); ok {
		return traceID
	}

	return ""
}
