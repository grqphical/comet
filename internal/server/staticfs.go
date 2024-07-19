package server

import (
	"comet/internal"
	"comet/internal/config"
	"comet/internal/logging"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
)

type StaticFS struct {
	directory http.Dir
	backend   *config.Backend
}

func newStaticFS(backend *config.Backend) *StaticFS {
	stat, err := os.Stat(backend.Directory)
	if err != nil {
		logging.LogCritical("unable to open staticfs directory")
	} else if !stat.IsDir() {
		logging.LogCritical("staticfs directory is not a directory")
	}

	return &StaticFS{
		http.Dir(backend.Directory),
		backend,
	}
}

func (s *StaticFS) HandleRequest(w http.ResponseWriter, r *http.Request) int {
	fileName, _ := removeFilterPrefix(s.backend.RouteFilter, r.URL.RequestURI())

	mimeType := mime.TypeByExtension(filepath.Ext(fileName))

	file, err := s.directory.Open(fileName)
	if err != nil {
		sendError(404, w)
		return 404
	}

	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Server", fmt.Sprintf("Comet/%s Go/1.22.5", internal.Version))

	_, err = io.Copy(w, file)
	if err != nil {
		sendError(500, w)
		return 500
	}

	return 200
}

// since there is no health to check this function is a no op
func (s *StaticFS) CheckHealth() {

}
