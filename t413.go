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
	synPrefix    = []byte("/syn/")
	staticPrefix = []byte("/static/")
	icoPrefix    = []byte("/ico/")
	infosPrefix  = []byte("/infos/")
	scanPrefix   = []byte("/scan/")
	artPrefix    = []byte("/art/")
	favicoPrefix = []byte("/favicon.ico")
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
	fasthttp.ServeFileUncompressed(ctx, filepath)
}

func sendLogo(ctx *fasthttp.RequestCtx) {
	ctx.SetContentType("image/png")
	sendBinary(ctx, "./icons/tmdb.png")
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
	sendLogo(ctx)
}

func icoserve(ctx *fasthttp.RequestCtx, DB *MovieDB.MDB) {
	query := strings.Split(string(ctx.Path()[1:]), "/")
	ico := fmt.Sprintf("./icons/%s", query[1])
	ctx.SetContentType("image/png")
	sendBinary(ctx, ico)
}

func staticserve(ctx *fasthttp.RequestCtx, DB *MovieDB.MDB) {
	path := strings.Replace(string(ctx.Path()), "/static", "", 1)
	// file := fmt.Sprintf("./icons/%s.png", query[1])
	sendBinary(ctx, path)
}

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
	case bytes.HasPrefix(path, staticPrefix):
		staticserve(ctx, DB)
	case bytes.HasPrefix(path, icoPrefix):
		icoserve(ctx, DB)
	case bytes.HasPrefix(path, scanPrefix):
		scandir(ctx)
	case bytes.HasPrefix(path, artPrefix):
		artworks(ctx, DB)
	case bytes.HasPrefix(path, favicoPrefix):
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
