package capture

import (
	dh "dnsrebinder/http"
	"fmt"
	"github.com/golang/glog"
	"github.com/gorilla/handlers"
	"html/template"
	"io/ioutil"
	h "net/http"
	"os"
	"path"
	"strconv"
	"strings"
)

var ReadBufSize int = 4096
var WriteBufSize int = 4096
var CaptureRoot string
var HTTPState map[string]int

var server *h.Server

var t *template.Template
var js template.HTML

func serveHttp(bind string) {
	mux := h.NewServeMux()

	mux.HandleFunc("/", root)
	server = &h.Server{
		Addr:    bind,
		Handler: handlers.LoggingHandler(os.Stdout, mux),
	}
	panic(server.ListenAndServe())
}

type templateInfo struct {
	R             *h.Request
	PrefetchCount []int
	Script        template.HTML
}

func root(w h.ResponseWriter, r *h.Request) {
	props := r.URL.Query()
	domain := strings.Split(r.Host, ":")[0]
	he := w.Header()
	if props.Get("noConnectionClose") == "" {
		he.Add("Connection", "close")
	}
	if props.Get("noXDNSPrefetchControl") == "" {
		he.Add("X-DNS-Prefetch-Control", "on")
	}
	if props.Get("noCacheControl") == "" {
		he.Add("Cache-Control", "private, no-cache, no-store, must-revalidate")
	}

	var info templateInfo
	info.R = r
	info.Script = js
	prefetches := props.Get("prefetchCount")
	if prefetches == "" {
		info.PrefetchCount = make([]int, 0, 0)
	} else {
		x, err := strconv.Atoi(prefetches)
		if err != nil {
			h.Error(w, "Internal Server Error", 500)
			return
		}
		info.PrefetchCount = make([]int, x)
	}

	err := t.Execute(w, &info)
	if err != nil {
		glog.Warningln("Capture Template error", err)
	}
	glog.Infoln("Captured for domain:", domain)

	dh.BroadcastToListeners(fmt.Sprintf("Payload HTTP request on %s", r.Host))

	HTTPState[domain] = 1
}

func StartHTTPCaptureServer(bind string) {
	t = template.Must(template.ParseFiles(path.Join(CaptureRoot, "index.html")))
	b, err := ioutil.ReadFile(path.Join(CaptureRoot, "capture.js"))
	if err != nil {
		panic("Failed to open capture js")
	}
	js = template.HTML("<script>" + string(b) + "</script>")
	go serveHttp(bind)
}
