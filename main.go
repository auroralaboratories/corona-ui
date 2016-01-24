package main

import (
    "fmt"
    "net"
    "os"
    "strconv"
    "strings"

    "github.com/auroralaboratories/corona-ui/util"
    "github.com/codegangsta/cli"
    "github.com/ghetzel/diecast/diecast"
    "github.com/auroralaboratories/go-gtk/gtk"
    "github.com/auroralaboratories/go-webkit/webkit"
    log "github.com/Sirupsen/logrus"
)

const (
    DEFAULT_UI_TEMPLATE_PATH = `ui/src`
    DEFAULT_UI_STATIC_PATH   = `ui/static`
    DEFAULT_UI_CONFIG_FILE   = `ui/config.yml`
)

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

        dc              := diecast.NewServer()
        dc.Address       = c.String(`address`)
        dc.TemplatePath  = c.String(`template-dir`)
        dc.StaticPath    = c.String(`static-dir`)
        dc.ConfigPath    = c.String(`ui-config`)

        if c.Bool(`quiet`) {
            dc.LogLevel = `quiet`
        }else{
            dc.LogLevel = c.String(`log-level`)
        }

        if port := c.Int(`port`); port == 0 {
            if listener, err := net.Listen(`tcp`, fmt.Sprintf("%s:%d", dc.Address, 0)); err == nil {
                parts := strings.SplitN(listener.Addr().String(), `:`, 2)

                if len(parts) == 2 {
                    if v, err := strconv.ParseUint(parts[1], 10, 32); err == nil {
                        dc.Port = int(v)
                    }else{
                        log.Fatalf("Unable to allocate UI server port: %v", err)
                    }
                }else{
                    log.Fatalf("Unable to allocate UI server port")
                }

                if err := listener.Close(); err != nil {
                    log.Fatalf("Failed to close ephemeral port allocator: %v", err)
                }
            }
        }else{
            dc.Port = c.Int(`port`)
        }

        if err := dc.Initialize(); err == nil {
            go func(){
                log.Infof("Diecast UI server at http://%s:%d", dc.Address, dc.Port)
                dc.Serve()
            }()
        }else{
            log.Fatalf("Failed to initialize UI server: %v", err)
        }

        gtk.Init(nil)
        window := gtk.NewWindow(gtk.WINDOW_TOPLEVEL)
        window.SetTitle(util.ApplicationName)
        // window.SetOpacity(0.5)
        window.Connect(`destroy`, gtk.MainQuit)

        scr := window.GetScreen()
        log.Infof("Screen for window: %+v (%T)", scr, scr)

        if cmap := window.GetColormap(); cmap.GColormap != nil {
            log.Infof("Setting colormap: %+v (%T)", cmap, cmap)
            window.SetColormap(cmap)
        }

        vbox := gtk.NewVBox(false, 1)

        webview := webkit.NewWebView()
        // webview.SetTransparent(true)

        webview.LoadUri(fmt.Sprintf("http://%s:%d", dc.Address, dc.Port))

        webview.Connect(`resource-load-finished`, func() {
            log.Infof("Loaded %s", webview.GetUri())
        })

        vbox.Add(webview)

        // entry.Connect("activate", func() {

        // })
        // button := gtk.NewButtonWithLabel("load String")
        // button.Clicked(func() {
        //     webview.LoadString("hello Go GTK!", "text/plain", "utf-8", ".")
        // })
        // vbox.PackStart(button, false, false, 0)

        // button = gtk.NewButtonWithLabel("load HTML String")
        // button.Clicked(func() {
        //     webview.LoadHtmlString(HTML_STRING, ".")
        // })
        // vbox.PackStart(button, false, false, 0)

        // button = gtk.NewButtonWithLabel("Google Maps")
        // button.Clicked(func() {
        //     webview.LoadHtmlString(MAP_EMBED, ".")
        // })
        // vbox.PackStart(button, false, false, 0)

        window.Add(vbox)
        window.SetSizeRequest(600, 600)
        window.ShowAll()

        proxy := os.Getenv(`HTTP_PROXY`)
        if len(proxy) > 0 {
            soup_uri := webkit.SoupUri(proxy)
            webkit.GetDefaultSession().Set(`proxy-uri`, soup_uri)
            soup_uri.Free()
        }

        gtk.Main()
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
            Value:  `127.0.0.1`,
        },
        cli.IntFlag{
            Name:   `port, p`,
            Usage:  `The port the diecast UI server should listen on`,
            Value:  0,
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
            Name:   `ui-config, c`,
            Usage:  `The path to the UI configuration file`,
            Value:  DEFAULT_UI_CONFIG_FILE,
        },
    }

    app.Run(os.Args)
}