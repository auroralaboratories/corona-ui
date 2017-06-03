package main

//go:generate esc -o util/static.go -pkg util -prefix embed embed

import (
	"fmt"
	"github.com/auroralaboratories/corona-ui/util"
	"github.com/ghetzel/diecast"
	"github.com/husobee/vestigo"
	"github.com/urfave/negroni"
	"net/http"
	"strings"
)

const (
	DEFAULT_UI_SERVER_ADDR = `127.0.0.1:19059`
)

type Server struct {
	Address    string  `json:"address"`
	RootPath   string  `json:"root"`
	ConfigPath string  `json:"config_path"`
	EmbedPath  string  `json:"embed_path"`
	Window     *Window `json:"-"`
}

func (self *Server) Serve() error {
	var embedFS http.FileSystem

	if self.EmbedPath == `` {
		embedFS = util.FS(false)
		log.Debugf("Using embedded")
	} else {
		embedFS = http.Dir(self.EmbedPath)
	}

	server := negroni.New()
	mux := http.NewServeMux()
	appRenderer := diecast.NewServer(`/`, `*.html`, `*.js`, `*.css`)
	appRenderer.VerifyFile = `/index.html`
	appRenderer.SetFileSystem(http.Dir(self.RootPath))

	// URLs under "/corona/{!api}*" are handled by the embedded filesystem
	appRenderer.SetMounts([]diecast.Mount{
		&diecast.FileMount{
			MountPoint: `/corona`,
			FileSystem: embedFS,
		},
	})

	if err := appRenderer.Initialize(); err != nil {
		return err
	}

	api := vestigo.NewRouter()
	self.setupApiRoutes(api)

	mux.Handle(`/corona/api/`, api)
	mux.Handle(`/`, appRenderer)

	server.UseHandler(mux)
	server.Use(NewRequestLogger())

	log.Debugf("Running API server at %s", self.Address)
	server.Run(self.Address)
	return nil
}

func (self *Server) GetURL() string {
	addr := self.Address

	if strings.HasPrefix(addr, `:`) {
		addr = `127.0.0.1` + addr
	}

	return fmt.Sprintf("http://%s", addr)
}

func (self *Server) setupApiRoutes(api *vestigo.Router) {
	api.Get(`/corona/api/status`, func(w http.ResponseWriter, req *http.Request) {
		util.Respond(w, map[string]interface{}{
			`version`: util.ApplicationVersion,
		})
	})
}
