package main

import (
	"fmt"
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
	Window     *Window `json:"-"`
}

func NewServer() *Server {
	return &Server{
		Address: DEFAULT_UI_SERVER_ADDR,
	}
}

func (self *Server) Serve() error {
	server := negroni.New()
	router := vestigo.NewRouter()
	ui := diecast.NewServer(self.RootPath, `*.html`)

	if err := ui.Initialize(); err != nil {
		return err
	}

	// routes not registered below will fallback to the UI server
	vestigo.CustomNotFoundHandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ui.ServeHTTP(w, req)
	})

	server.UseHandler(router)

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
