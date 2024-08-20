package server

import (
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

	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			handleRoot(w, r, staticFS)
		} else {
			fileServer.ServeHTTP(w, r)
		}
	}))

	mux.HandleFunc("/connect", handleConnect)
	mux.HandleFunc("/status", handleStatus)
	mux.HandleFunc("/info", handleInfo)
	mux.HandleFunc("/terminal", handleTerminal)

	return &http.Server{
		Addr:    ":8611",
		Handler: mux,
	}
}

func handleRoot(w http.ResponseWriter, r *http.Request, staticFS fs.FS) {
	var filename string
	if client.IsConnected() {
		filename = "index.html"
	} else {
		filename = "login.html"
	}

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
