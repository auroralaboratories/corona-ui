package main

import (
    "fmt"
    "net"
    "os"
    "strconv"
    "strings"
    "unsafe"

    "github.com/auroralaboratories/corona-ui/util"
    "github.com/codegangsta/cli"
    "github.com/ghetzel/diecast/diecast"
    "github.com/auroralaboratories/go-gtk/gtk"
    "github.com/auroralaboratories/go-gtk/glib"
    "github.com/auroralaboratories/go-cairo"
    "github.com/auroralaboratories/go-webkit/webkit"
    log "github.com/Sirupsen/logrus"
)

const (
    DEFAULT_UI_TEMPLATE_PATH = `ui/src`
    DEFAULT_UI_STATIC_PATH   = `ui/static`
    DEFAULT_UI_CONFIG_FILE   = `ui/config.yml`
)

var useAlpha bool

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

        window  := gtk.NewWindow(gtk.WINDOW_TOPLEVEL)
        layout  := gtk.NewScrolledWindow(nil, nil)
        webview := webkit.NewWebView()
        webset  := webkit.NewWebSettings()

        window.Connect(`destroy`,         gtk.MainQuit)

        layout.Connect(`screen-changed`,  UpdateScreen)
        layout.Connect(`expose-event`,    ExposeEvent)

        webview.Connect(`screen-changed`, UpdateScreen)
        webview.Connect(`expose-event`,   ExposeEvent)

        webview.Connect(`resource-load-finished`, func() {
            log.Infof("Loaded %s", webview.GetUri())
        })

        window.SetAppPaintable(true)
        layout.SetAppPaintable(true)


        webset.Set("auto-load-images",                  true)
        webset.Set("auto-resize-window",                false)
        webset.Set("enable-plugins",                    true)
        webset.Set("enable-scripts",                    true)
        webset.Set("enable-accelerated-compositing",    true)
        webset.Set("enable-webgl",                      true)
        webset.Set("enable-webaudio",                   true)
        webset.Set("enable-file-access-from-file-uris", true)

        webview.SetSettings(webset)
        webview.SetTransparent(true)

        layout.Add(webview)
        window.Add(layout)

        webview.LoadUri(fmt.Sprintf("http://%s:%d", dc.Address, dc.Port))

        window.SetSizeRequest(600, 600)


        UpdateWidgetScreen(gtk.WidgetFromNative(unsafe.Pointer(window.ToNative())))

        window.ShowAll()

        // proxy := os.Getenv(`HTTP_PROXY`)
        // if len(proxy) > 0 {
        //     soup_uri := webkit.SoupUri(proxy)
        //     webkit.GetDefaultSession().Set(`proxy-uri`, soup_uri)
        //     soup_uri.Free()
        // }

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


func UpdateScreen(ctx *glib.CallbackContext) {
    if tgt := ctx.Target(); tgt != nil {
        widget := gtk.WidgetFromObject(tgt.(*glib.GObject))
        UpdateWidgetScreen(widget)
    }
}

func UpdateWidgetScreen(widget *gtk.Widget) {
    log.Debugf("screen-changed widget: %T %+v", widget, widget)

    screen := widget.GetScreen()

    if screen.IsComposited() {
        log.Infof("Compositing is enabled")
        useAlpha = true
        widget.SetColormap(screen.GetRGBAColormap())
    }else{
        log.Warnf("Compositing is disabled")
        useAlpha = false
        widget.SetColormap(screen.GetRGBColormap())
    }
}

func ExposeEvent(ctx *glib.CallbackContext) {
    if tgt := ctx.Target(); tgt != nil {
        widget := gtk.WidgetFromObject(tgt.(*glib.GObject))
        log.Debugf("expose-event widget: %T %+v", widget, widget)

        if window := widget.GetParentWindow(); window != nil {
            if drawable := window.GetDrawable(); drawable != nil {
                context := drawable.CairoCreate()
                target  := cairo.GetTarget(context)

                if surface := cairo.NewSurfaceFromC(target, context); surface != nil {
                    if useAlpha {
                        surface.SetSourceRGBA(0.0, 0.1647, 1.0, 0.5)
                    }else{
                        surface.SetSourceRGB(1.0, 1.0, 1.0)
                    }

                    surface.SetOperator(cairo.OPERATOR_SOURCE)
                    surface.Paint()
                    surface.Destroy()
                }else{
                    log.Debugf("expose-event: failed to create cairo surface of %+v", widget)
                }
            }else{
                log.Debugf("expose-event: failed to get drawable surface of %+v", widget)
            }
        }else{
            log.Debugf("expose-event: failed to get parent window of %+v", widget)
        }
    }
}