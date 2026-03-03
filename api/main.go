package main

import (
	"log"
	"log/slog"

	"github.com/spiritov/svelte-go-steam/db"
	"github.com/spiritov/svelte-go-steam/env"
	"github.com/spiritov/svelte-go-steam/internal"
	"github.com/spiritov/svelte-go-steam/slogger"
)

func main() {
	// env is a wrapper around the `godotenv` library
	err := env.Load("SITE_ENV")
	if err != nil {
		slog.Error("error loading .env", "error", err)
		log.Fatal()
	}

	env.Require(
		"API_HTTP_ADDRESS",
		"API_OID_REALM",
		"API_SESSION_COOKIE_SECURE",
		"API_SESSION_TOKEN_SECRET",
		"API_STEAM_API_KEY",
		"API_SLOG_LEVEL",
		"API_SLOG_MODE",
		"API_HTTP_ADDRESS",
		"API_HTTPLOG_MODE",
		"API_HTTPLOG_LEVEL",
		"API_HTTPLOG_CONCISE",
		"API_HTTPLOG_REQUEST_HEADERS",
		"API_HTTPLOG_RESPONSE_HEADERS",
		"API_HTTPLOG_REQUEST_BODIES",
		"API_HTTPLOG_RESPONSE_BODIES",
	)

	err = slogger.Setup()
	if err != nil {
		log.Fatal(err)
	}
	slog.SetDefault(slogger.Logger)

	database := db.OpenDB("./db/jump.db")
	defer database.Close()

	slog.Info("sqlite db up")

	address := env.GetString("API_HTTP_ADDRESS")
	internal.ServeAPI(address)
}
