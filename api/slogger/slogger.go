package slogger

import (
	"log/slog"
	"os"

	"github.com/rotisserie/eris"
	"github.com/spiritov/svelte-go-steam/env"
)

var Logger *slog.Logger

var SlogLevelMap = map[string]slog.Level{
	"Debug": slog.LevelDebug,
	"Info":  slog.LevelInfo,
	"Warn":  slog.LevelWarn,
	"Error": slog.LevelError,
}

func Setup() error {
	logLevel, matchedErr := env.GetMapped("API_SLOG_LEVEL", SlogLevelMap)
	if matchedErr != nil {
		return matchedErr
	}

	handlerOptions := &slog.HandlerOptions{Level: logLevel}

	slogMode := env.GetString("API_SLOG_MODE")
	switch slogMode {
	case "Text":
		Logger = slog.New(slog.NewTextHandler(os.Stdout, handlerOptions))
	case "JSON":
		Logger = slog.New(slog.NewJSONHandler(os.Stdout, handlerOptions))
	default:
		return eris.Errorf("invalid value for API_SLOG_MODE: %s", slogMode)
	}

	return nil
}
