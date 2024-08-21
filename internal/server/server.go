package server

import (
	"github.com/rs/cors"
	"github.com/rxxuzi/sirius/internal/client"
	"github.com/rxxuzi/sirius/internal/static"
	"io"
	"io/fs"
	"net/http"
)

func NewServer() *http.Server {
	mux := http.NewServeMux()

	staticFS := static.GetFS()
	fileServer := http.FileServer(http.FS(staticFS))
	handler := cors.Default().Handler(mux)

	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			handleRoot(w, r, staticFS)
		} else {
			fileServer.ServeHTTP(w, r)
		}
	}))

	mux.HandleFunc("/connect", handleConnect)
	mux.HandleFunc("/status", handleStatus)
	mux.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
		serveHTMLFile(w, r, staticFS, "info.html")
	})
	mux.HandleFunc("/api/static-info", handleStaticInfo)
	mux.HandleFunc("/api/live-info", handleLiveInfo)

	mux.HandleFunc("/terminal", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Upgrade") == "websocket" {
			handleTerminal(w, r)
		} else {
			serveHTMLFile(w, r, staticFS, "terminal.html")
		}
	})

	mux.HandleFunc("/code", func(w http.ResponseWriter, r *http.Request) {
		serveHTMLFile(w, r, staticFS, "code.html")
	})

	mux.HandleFunc("/ws", handleTerminal)

	return &http.Server{
		Addr:    ":8611",
		Handler: handler,
	}
}

func handleRoot(w http.ResponseWriter, r *http.Request, staticFS fs.FS) {
	var filename string
	if client.IsConnected() {
		filename = "index.html"
	} else {
		filename = "login.html"
	}

	serveHTMLFile(w, r, staticFS, filename)
}

func serveHTMLFile(w http.ResponseWriter, r *http.Request, staticFS fs.FS, filename string) {
	file, err := staticFS.Open(filename)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.ServeContent(w, r, filename, stat.ModTime(), file.(io.ReadSeeker))
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	if client.IsConnected() {
		w.Write([]byte("Connected"))
	} else {
		w.Write([]byte("Not connected"))
	}
}
