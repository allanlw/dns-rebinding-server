package http

import (
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"github.com/gorilla/handlers"
	"github.com/gorilla/websocket"
	"io/ioutil"
	h "net/http"
	"os"
	"path"
	"sync"
)

var ReadBufSize int = 4096
var WriteBufSize int = 4096
var WebRoot string

var records map[string][]byte = make(map[string][]byte)
var recordsLock sync.Mutex

func simpleFileServer(f string) func(w h.ResponseWriter, r *h.Request) {
	return func(w h.ResponseWriter, r *h.Request) {
		h.ServeFile(w, r, path.Join(WebRoot, f))
	}
}

func serveHttp(bind string, config interface{}) {
	h.HandleFunc("/ws", ws)
	h.HandleFunc("/record", record)
	h.HandleFunc("/server-settings.json", func(w h.ResponseWriter, r *h.Request) {
		w.Header().Add("Content-Type", "application/json")
		e := json.NewEncoder(w)
		e.Encode(config)
	})

	h.HandleFunc("/", simpleFileServer("index.html"))
	h.HandleFunc("/attack", simpleFileServer("attack.html"))
	h.HandleFunc("/attack.js", simpleFileServer("attack.js"))
	h.HandleFunc("/demo/step-by-step", simpleFileServer("demo/step-by-step.html"))
	h.HandleFunc("/demo/step-by-step.js", simpleFileServer("demo/step-by-step.js"))
	h.HandleFunc("/demo/full", simpleFileServer("demo/full.html"))
	h.HandleFunc("/demo/full.js", simpleFileServer("demo/full.js"))

	panic(h.ListenAndServe(bind, handlers.LoggingHandler(os.Stdout, h.DefaultServeMux)))
}

func record(w h.ResponseWriter, r *h.Request) {
	id := r.URL.Query().Get("id")
	var a []byte
	var err error

	recordsLock.Lock()
	defer recordsLock.Unlock()

	switch r.Method {
	case "GET":
		break
	case "POST":
		a, err = ioutil.ReadAll(r.Body)
		if err != nil {
			glog.Infoln("Failed to read request body", err)
			w.WriteHeader(500)
			return
		}
		glog.Info("Got POST for id: %s, creating record.\n", id)
		records[id] = a
	default:
		w.WriteHeader(404)
		return
	}
	a, ok := records[id]
	if !ok {
		glog.Infof("Got GET for id: %s, but don't have a record for it.\n", id)
		w.WriteHeader(404)
		return
	}
	w.WriteHeader(200)
	_, err = w.Write(a)
	if err != nil {
		glog.Warningln("Failed to write value to response body", err)
		return
	}
}

func ws(w h.ResponseWriter, r *h.Request) {
	conn, err := websocket.Upgrade(w, r, w.Header(), ReadBufSize, WriteBufSize)
	if err != nil {
		h.Error(w, "websocket error", h.StatusBadRequest)
		fmt.Println("Websocket error", err)
	}

	go provideWs(conn)
}

func provideWs(conn *websocket.Conn) {
	wrapper := wrapWebSocket(conn)

	addListener(wrapper)
}

func StartHTTPServer(bind string, config interface{}) {
	go serveHttp(bind, config)
}
