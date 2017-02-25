package main

import (
	"github.com/auroralaboratories/corona-ui/util"
	"github.com/ghetzel/cli"
	"github.com/ghodss/yaml"
	"github.com/op/go-logging"
	"io/ioutil"
	"os"
)

const (
	DEFAULT_UI_CONFIG_FILE = `config.yml`
)

var log = logging.MustGetLogger(`main`)

var useAlpha bool = false
var server Server
var config Config = GetDefaultConfig()

func main() {
	app := cli.NewApp()
	app.Name = util.ApplicationName
	app.Usage = util.ApplicationSummary
	app.Version = util.ApplicationVersion
	app.EnableBashCompletion = false

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
		cli.StringFlag{
			Name:  `config, c`,
			Usage: `The path to the configuration file`,
			Value: DEFAULT_UI_CONFIG_FILE,
		},
		cli.BoolFlag{
			Name:  `server-only`,
			Usage: `Only start the UI server (and skip creating and showing the window)`,
		},
	}

	app.Before = func(c *cli.Context) error {
		logging.SetFormatter(logging.MustStringFormatter(`%{color}%{level:.4s}%{color:reset}[%{id:04d}] %{message}`))

		if level, err := logging.LogLevel(c.String(`log-level`)); err == nil {
			logging.SetLevel(level, ``)
		} else {
			return err
		}

		return nil
	}

	app.Action = func(c *cli.Context) {
		log.Infof("%s v%s started at %s", util.ApplicationName, util.ApplicationVersion, util.StartedAt)

		if data, err := ioutil.ReadFile(c.String(`config`)); err == nil {
			log.Debugf("Default Configuration: %+v", config)

			if err := yaml.Unmarshal(data, &config); err == nil {
				log.Infof("Successfully loaded configuration file: %s", c.String(`config`))
				log.Debugf("Configuration: %+v", config)
			}
		}

		//  start the UI server in the background
		if err := startUiServer(c); err != nil {
			log.Fatalf("%v", err)
		}

		//  setup and show the window
		if c.Bool(`server-only`) {
			log.Noticef("Started in server-only mode; no GUI elements will be shown.")
			select {}
		} else {
			window := NewWindow(&server)
			server.Window = window

			if len(c.Args()) > 0 {
				window.URI = c.Args()[0]
			}

			if err := window.Initialize(&config.Window); err == nil {
				if err := window.Show(); err != nil {
					log.Fatalf("%v", err)
				}
			} else {
				log.Fatalf("Failed to initialize window: %v", err)
			}
		}
	}

	app.Run(os.Args)
}

func startUiServer(c *cli.Context) error {
	server = config.Server
	server.Address = c.String(`address`)
	server.ConfigPath = c.String(`config`)

	if c.NArg() > 0 {
		server.RootPath = c.Args().First()
	}

	go func() {
		log.Infof("UI server at http://%s", server.Address)
		server.Serve()
	}()

	return nil
}
