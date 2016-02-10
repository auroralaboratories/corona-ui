package main

import (
    "fmt"
    "io/ioutil"
    "os"

    "github.com/auroralaboratories/corona-ui/util"
    "github.com/codegangsta/cli"
    "gopkg.in/yaml.v2"
    log "github.com/Sirupsen/logrus"
)



var useAlpha bool    = false
var server   *Server
var config   Config  = GetDefaultConfig()

func main(){
    app                      := cli.NewApp()
    app.Name                  = util.ApplicationName
    app.Usage                 = util.ApplicationSummary
    app.Version               = util.ApplicationVersion
    app.EnableBashCompletion  = false
    app.Action                = func(c *cli.Context) {
        if c.Bool(`quiet`) {
            util.ParseLogLevel(`quiet`)
        }else{
            util.ParseLogLevel(c.String(`log-level`))
        }

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
        window := NewWindow(server)

        if len(c.Args()) > 0 {
            window.URI = c.Args()[0]
        }

        if err := window.Initialize(&config.Window); err == nil {
            if err := window.Show(); err != nil {
                log.Fatalf("%v", err)
            }
        }else{
            log.Fatalf("Failed to initialize window: %v", err)
        }
    }

    app.Flags = []cli.Flag{
        cli.StringFlag{
            Name:   `log-level, L`,
            Usage:  `Level of log output verbosity`,
            Value:  `info`,
            EnvVar: `LOGLEVEL`,
        },
        cli.BoolFlag{
            Name:   `quiet, q`,
            Usage:  `Don't print any log output to standard error`,
        },
        cli.StringFlag{
            Name:   `address, a`,
            Usage:  `The address the diecast UI server should listen on`,
            Value:  DEFAULT_UI_SERVER_ADDR,
        },
        cli.IntFlag{
            Name:   `port, p`,
            Usage:  `The port the diecast UI server should listen on`,
            Value:  DEFAULT_UI_SERVER_PORT,
        },
        cli.StringFlag{
            Name:   `template-dir, T`,
            Usage:  `The directory containing the UI template definitions`,
            Value:  DEFAULT_UI_TEMPLATE_PATH,
        },
        cli.StringFlag{
            Name:   `static-dir, S`,
            Usage:  `The directory containing the UI static content`,
            Value:  DEFAULT_UI_STATIC_PATH,
        },
        cli.StringFlag{
            Name:   `config, c`,
            Usage:  `The path to the configuration file`,
            Value:  DEFAULT_UI_CONFIG_FILE,
        },
    }

    app.Run(os.Args)
}

func startUiServer(c *cli.Context) error {
    server               = NewServer()
    server.Address       = c.String(`address`)
    server.Port          = c.Int(`port`)
    server.TemplatePath  = c.String(`template-dir`)
    server.StaticPath    = c.String(`static-dir`)
    server.ConfigPath    = c.String(`config`)

    if c.Bool(`quiet`) {
        server.LogLevel = `quiet`
    }else{
        server.LogLevel = c.String(`log-level`)
    }

    if err := server.Initialize(); err == nil {
        go func(){
            log.Infof("UI server at http://%s:%d", server.Address, server.Port)
            server.Serve()
        }()
    }else{
        return fmt.Errorf("Failed to initialize UI server: %v", err)
    }

    return nil
}
