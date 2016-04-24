package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/auroralaboratories/corona-ui/util"
	"github.com/codegangsta/cli"
	"os"
)

var server *Server

func main() {
	app := cli.NewApp()
	app.Name = util.ApplicationName
	app.Usage = util.ApplicationSummary
	app.Version = util.ApplicationVersion
	app.EnableBashCompletion = false

	app.Action = func(c *cli.Context) {
		if c.Bool(`quiet`) {
			util.ParseLogLevel(`quiet`)
		} else {
			util.ParseLogLevel(c.String(`log-level`))
		}

		log.Infof("%s v%s started at %s", util.ApplicationName, util.ApplicationVersion, util.StartedAt)

		//  start the UI server in the background
		if err := startUiServer(c); err != nil {
			log.Fatalf("%v", err)
		}

		log.Warnf("Started in server-only mode; no GUI elements will be shown.")
		select {}
	}

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   `log-level, L`,
			Usage:  `Level of log output verbosity`,
			Value:  `info`,
			EnvVar: `LOGLEVEL`,
		},
		cli.BoolFlag{
			Name:  `quiet, q`,
			Usage: `Don't print any log output to standard error`,
		},
		cli.StringFlag{
			Name:  `address, a`,
			Usage: `The address the diecast UI server should listen on`,
			Value: DEFAULT_UI_SERVER_ADDR,
		},
		cli.IntFlag{
			Name:  `port, p`,
			Usage: `The port the diecast UI server should listen on`,
			Value: 28419,
		},
		cli.StringFlag{
			Name:  `embed-dir`,
			Usage: `The directory containing embedded assets`,
			Value: DEFAULT_UI_EMBED_PATH,
		},
		cli.StringFlag{
			Name:  `embed-route`,
			Usage: `The HTTP path that will be used to serve embedded assets`,
			Value: DEFAULT_UI_EMBED_ROUTE,
		},
		cli.StringFlag{
			Name:  `template-dir, T`,
			Usage: `The directory containing the UI template definitions`,
			Value: DEFAULT_UI_TEMPLATE_PATH,
		},
		cli.StringFlag{
			Name:  `static-dir, S`,
			Usage: `The directory containing the UI static content`,
			Value: DEFAULT_UI_STATIC_PATH,
		},
		cli.BoolFlag{
			Name:  `server-only`,
			Usage: `Only start the UI server (and skip creating and showing the window)`,
		},
	}

	app.Run(os.Args)
}

func startUiServer(c *cli.Context) error {
	server = NewServer()
	server.Address = c.String(`address`)
	server.ConfigPath = c.String(`config`)
	server.EmbedPath = c.String(`embed-dir`)
	server.EmbedRoute = c.String(`embed-route`)
	server.Port = c.Int(`port`)
	server.StaticPath = c.String(`static-dir`)
	server.TemplatePath = c.String(`template-dir`)

	if c.Bool(`quiet`) {
		server.LogLevel = `quiet`
	} else {
		server.LogLevel = c.String(`log-level`)
	}

	if err := server.Initialize(); err == nil {
		go func() {
			log.Infof("UI server at http://%s:%d", server.Address, server.Port)
			server.Serve()
		}()
	} else {
		return fmt.Errorf("Failed to initialize UI server: %v", err)
	}

	return nil
}
