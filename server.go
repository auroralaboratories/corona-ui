package main

import (
	"fmt"
	"github.com/ghetzel/diecast/diecast"
	"net"
	"strconv"
	"strings"
)

const (
	DEFAULT_UI_SERVER_ADDR     = `127.0.0.1`
	DEFAULT_UI_SERVER_PORT     = 0
	DEFAULT_UI_TEMPLATE_PATH   = `src`
	DEFAULT_UI_STATIC_PATH     = `static`
	DEFAULT_UI_CONFIG_FILE     = `config.yml`
	DEFAULT_UI_EMBED_PATH      = `embedded`
	DEFAULT_UI_EMBED_ROUTE     = `/corona/`
	DEFAULT_UI_EMBED_API_ROUTE = `/corona/api/`
)

type Server struct {
	Address      string
	Port         int
	TemplatePath string
	ConfigPath   string
	StaticPath   string
	LogLevel     string
	EmbedPath    string
	EmbedRoute   string

	dc *diecast.Server
}

func NewServer() *Server {
	return &Server{
		Address:      DEFAULT_UI_SERVER_ADDR,
		ConfigPath:   DEFAULT_UI_CONFIG_FILE,
		Port:         DEFAULT_UI_SERVER_PORT,
		StaticPath:   DEFAULT_UI_STATIC_PATH,
		TemplatePath: DEFAULT_UI_TEMPLATE_PATH,
		EmbedPath:    DEFAULT_UI_EMBED_PATH,
		EmbedRoute:   DEFAULT_UI_EMBED_ROUTE,
	}
}

func (self *Server) Initialize() error {
	self.dc = diecast.NewServer()
	self.dc.Address = self.Address
	self.dc.TemplatePath = self.TemplatePath
	self.dc.StaticPath = self.StaticPath
	self.dc.ConfigPath = self.ConfigPath
	self.dc.LogLevel = self.LogLevel

	self.registerHandlers()

	if self.Port == 0 {
		if listener, err := net.Listen(`tcp`, fmt.Sprintf("%s:%d", self.dc.Address, 0)); err == nil {
			parts := strings.SplitN(listener.Addr().String(), `:`, 2)

			if len(parts) == 2 {
				if v, err := strconv.ParseUint(parts[1], 10, 32); err == nil {
					self.dc.Port = int(v)
					self.Port = self.dc.Port
				} else {
					return fmt.Errorf("Unable to allocate UI server port: %v", err)
				}
			} else {
				return fmt.Errorf("Unable to allocate UI server port")
			}

			if err := listener.Close(); err != nil {
				return fmt.Errorf("Failed to close ephemeral port allocator: %v", err)
			}
		}
	} else {
		self.dc.Port = self.Port
	}

	return self.dc.Initialize()
}

func (self *Server) SetPayload(key string, value interface{}) {
	if self.dc != nil {
		if v, ok := self.dc.Payload[`corona`]; ok {
			switch v.(type) {
			case map[string]interface{}:
				coronaPayload := v.(map[string]interface{})
				coronaPayload[key] = value
			}
		} else {
			self.dc.Payload[`corona`] = map[string]interface{}{
				key: value,
			}
		}
	}
}

func (self *Server) Serve() error {
	if self.dc != nil {
		return self.dc.Serve()
	} else {
		return fmt.Errorf("Cannot start uninitialized server")
	}
}
