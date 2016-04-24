package main

import (
	"bytes"
	"fmt"
	"github.com/auroralaboratories/corona-ui/util"
	"github.com/ghetzel/diecast/diecast"
	"io"
	"mime"
	"net/http"
	"path"
	"strings"
)

func (self *Server) registerHandlers() {
	self.dc.HandleFuncs = append(self.dc.HandleFuncs, diecast.HandleFunc{
		Pattern: self.EmbedRoute,
		HandleFunc: func(w http.ResponseWriter, req *http.Request) {
			var reader io.Reader

			filepath := strings.TrimPrefix(req.URL.Path, self.EmbedRoute)

			if self.EmbedPath == `embedded` {
				if data, err := util.Asset(filepath); err == nil {
					reader = bytes.NewBuffer(data)
				} else {
					http.Error(w, fmt.Errorf("Cannot locate file '%s': %v", filepath, err.Error()).Error(), http.StatusNotFound)
					return
				}
			} else {
				var fs http.FileSystem
				fs = http.Dir(self.EmbedPath)

				if file, err := fs.Open(filepath); err == nil {
					reader = file
				} else {
					http.Error(w, fmt.Errorf("Cannot locate file '%s': %v", filepath, err.Error()).Error(), http.StatusNotFound)
					return
				}
			}

			contentType := mime.TypeByExtension(path.Ext(filepath))

			if contentType == `` {
				contentType = `text/plain`
			}

			w.Header().Set(`Content-Type`, contentType)

			if _, err := io.Copy(w, reader); err != nil {
				http.Error(w, fmt.Errorf("Unable to read file '%s': %v", filepath, err.Error()).Error(), http.StatusInternalServerError)
			}
		},
	})
}
