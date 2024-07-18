package server

import (
	"comet/internal/config"
	"comet/internal/logging"
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
	fileName, err := removeFilterPrefix(s.backend.RouteFilter, r.URL.RequestURI())
	if err != nil {
		logging.LogCritical("invalid route filter")
	}

	mimeType := mime.TypeByExtension(filepath.Ext(fileName))

	file, err := s.directory.Open(fileName)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return 404
	}

	w.Header().Set("Content-Type", mimeType)

	_, err = io.Copy(w, file)
	if err != nil {
		http.Error(w, "unable to read file data", http.StatusInternalServerError)
		return 500
	}

	return 200
}

// since there is no health to check this function is a no op
func (s *StaticFS) CheckHealth() {

}
