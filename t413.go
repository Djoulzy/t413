package main

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Djoulzy/MovieDB"
	"github.com/Djoulzy/ScanDir"
	"github.com/Djoulzy/Tools/clog"
	"github.com/Djoulzy/Tools/config"
	"github.com/valyala/fasthttp"
)

var (
	synPrefix   = []byte("/syn/")
	infosPrefix = []byte("/infos/")
	scanPrefix  = []byte("/scan/")
	artPrefix   = []byte("/art/")
	icoPrefix   = []byte("/favicon.ico")
)

type Globals struct {
	LogLevel     int
	StartLogging bool
	HTTP_addr    string
	TMDB_Key     string
	CacheDir     string
	PrefixDir    string
}

type AppConfig struct {
	Globals
}

var appConfig *AppConfig
var myDB *MovieDB.MDB

func (A AppConfig) GetHTTPAddr() string {
	return A.HTTP_addr
}

func (A AppConfig) GetTMDBKey() string {
	return A.TMDB_Key
}

func (A AppConfig) GetCacheDir() string {
	return A.CacheDir
}

func (A AppConfig) GetPrefixDir() string {
	return A.PrefixDir
}

func handleError(ctx *fasthttp.RequestCtx, message string, status int) {
	ctx.SetStatusCode(status)
	fmt.Fprintf(ctx, "%s\n", message)
}

func sendBuffer(ctx *fasthttp.RequestCtx, buffer []byte) {
	ctx.Write(buffer)
}

func sendBinary(ctx *fasthttp.RequestCtx, filepath string) {
	fasthttp.ServeFile(ctx, filepath)
}

func sendLogo(ctx *fasthttp.RequestCtx) {
	sendBinary(ctx, "./tmdb.png")
}

func artworks(ctx *fasthttp.RequestCtx, DB *MovieDB.MDB) {
	query := strings.Split(string(ctx.Path()[1:]), "/")
	if len(query) == 3 {
		url, err := DB.GetArtwork(query[1], query[2])
		if err == nil {
			ctx.SetContentType("image/jpg")
			sendBinary(ctx, url)
			return
		}
	}
	ctx.SetContentType("image/png")
	sendLogo(ctx)
}

// func synopsys(ctx *fasthttp.RequestCtx, DB *MovieDB.MDB) {
// 	var year string
// 	query := strings.Split(string(ctx.Path()[1:]), "/")
// 	if len(query) < 2 {
// 		handleError(ctx, "Bad Request", http.StatusNotFound)
// 		return
// 	}
// 	if len(query) < 3 {
// 		year = ""
// 	} else {
// 		year = query[2]
// 	}
// 	url, err := DB.GetSynopsys(query[1], year)
// 	if err != nil {
// 		ctx.SetContentType("image/png")
// 		ctx.Write([]byte("n/a"))
// 	} else {
// 		ctx.SetContentType("text/html")
// 		sendBinary(ctx, url)
// 	}
// }

func movieinfos(ctx *fasthttp.RequestCtx, DB *MovieDB.MDB) {
	query := strings.Split(string(ctx.Path()[1:]), "/")
	if len(query) < 2 {
		handleError(ctx, "Bad Request", http.StatusNotFound)
		return
	}

	url, err := DB.GetMovieInfos(query[1])
	if err != nil {
		ctx.SetContentType("text/plain")
		ctx.Write([]byte("n/a"))
	} else {
		ctx.SetContentType("application/json")
		sendBuffer(ctx, url)
	}
}

func scandir(ctx *fasthttp.RequestCtx) {
	// orderby := string(ctx.QueryArgs().Peek("orderby"))
	query := strings.Split(string(ctx.Path()[1:]), "/")
	if len(query) < 2 {
		handleError(ctx, "Bad Request", http.StatusNotFound)
		return
	}

	orderby := string(ctx.QueryArgs().Peek("orderby"))
	asc := (string(ctx.QueryArgs().Peek("desc")) == "")
	pagenum, _ := strconv.Atoi(string(ctx.QueryArgs().Peek("p")))
	nbperpage, _ := strconv.Atoi(string(ctx.QueryArgs().Peek("nb")))

	rep := ScanDir.Start(appConfig, myDB, strings.Join(query[1:], "/"), orderby, asc, pagenum, nbperpage)
	ctx.SetContentType("application/json")
	sendBuffer(ctx, rep)
}

func action(ctx *fasthttp.RequestCtx, DB *MovieDB.MDB) {
	clog.Info("t413", "action", "GET %s", ctx.Path())
	ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")

	path := ctx.Path()
	switch {
	// case bytes.HasPrefix(path, synPrefix):
	// 	synopsys(ctx, DB)
	case bytes.HasPrefix(path, infosPrefix):
		movieinfos(ctx, DB)
	case bytes.HasPrefix(path, scanPrefix):
		scandir(ctx)
	case bytes.HasPrefix(path, artPrefix):
		artworks(ctx, DB)
	case bytes.HasPrefix(path, icoPrefix):
		handleError(ctx, "Not found", http.StatusNotFound)
		return
	}
}

func main() {
	appConfig = &AppConfig{
		Globals{
			LogLevel:     5,
			StartLogging: true,
			HTTP_addr:    "localhost:9999",
		},
	}

	config.Load("t413.ini", appConfig)
	clog.LogLevel = appConfig.LogLevel
	clog.StartLogging = appConfig.StartLogging

	// ScanDir.MakePrettyName("Transformers.The.Last.Knight.2017.MULTI.1080p.WEB-DL.H264.WwW.Zone-Telechargement.Ws.mkv")
	// ScanDir.MakePrettyName("Alibi.com (2017) 1080p TRUEFRENCH x264 DTS - JiHeff.mkv")
	// return

	myDB = MovieDB.Init(appConfig)

	clog.Info("t413", "Start", "HTTP Listening on %s", appConfig.HTTP_addr)
	err := fasthttp.ListenAndServe(appConfig.HTTP_addr, func(ctx *fasthttp.RequestCtx) { action(ctx, myDB) })
	if err != nil {
		clog.Fatal("t413", "Start", err)
	}
}
