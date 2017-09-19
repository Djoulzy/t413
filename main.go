package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Djoulzy/MovieDB"
	"github.com/Djoulzy/Tools/clog"
	"github.com/Djoulzy/Tools/config"
	"github.com/valyala/fasthttp"
)

type Globals struct {
	LogLevel     int
	StartLogging bool
	HTTP_addr    string
	TMDB_Key     string
	CacheDir     string
}

type AppConfig struct {
	Globals
}

func (A AppConfig) GetTMDBKey() string {
	return A.TMDB_Key
}

func (A AppConfig) GetCacheDir() string {
	return A.CacheDir
}

func handleError(ctx *fasthttp.RequestCtx, message string, status int) {
	ctx.SetStatusCode(status)
	fmt.Fprintf(ctx, "%s\n", message)
}

func action(ctx *fasthttp.RequestCtx, DB *MovieDB.MDB) {

	clog.Info("t413", "action", "GET %s", ctx.Path())
	path := ctx.Path()
	query := strings.Split(string(path[1:]), "/")

	if len(query) < 2 {
		handleError(ctx, "Bad Query", http.StatusNotFound)
		return
	}

	switch query[0] {
	case "favicon.ico":
		handleError(ctx, "Not found", http.StatusNotFound)
		return
	case "syn":
		DB.GetSynopsys(ctx, query)
	default:
		DB.GetArtwork(ctx, query)
	}
}

func main() {
	appConfig := &AppConfig{
		Globals{
			LogLevel:     5,
			StartLogging: true,
			HTTP_addr:    "localhost:9999",
		},
	}

	config.Load("t413.ini", appConfig)
	clog.LogLevel = appConfig.LogLevel
	clog.StartLogging = appConfig.StartLogging

	myDB := MovieDB.Init(appConfig)

	clog.Info("t413", "Start", "HTTP Listening on %s", appConfig.HTTP_addr)
	err := fasthttp.ListenAndServe(appConfig.HTTP_addr, func(ctx *fasthttp.RequestCtx) { action(ctx, myDB) })
	if err != nil {
		clog.Fatal("t413", "Start", err)
	}
}
