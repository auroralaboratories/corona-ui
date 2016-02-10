package main

import (
    "fmt"
    "io/ioutil"
    "net"
    "os"
    "strconv"
    "strings"
    // "unsafe"

    "github.com/auroralaboratories/corona-ui/util"
    "github.com/gotk3/gotk3/cairo"
    "github.com/gotk3/gotk3/gdk"
    "github.com/gotk3/gotk3/glib"
    "github.com/gotk3/gotk3/gtk"
    "github.com/auroralaboratories/go-webkit2/webkit2"
    "github.com/codegangsta/cli"
    "github.com/ghetzel/diecast/diecast"
    "gopkg.in/yaml.v2"
    log "github.com/Sirupsen/logrus"
)

const (
    DEFAULT_UI_TEMPLATE_PATH = `src`
    DEFAULT_UI_STATIC_PATH   = `static`
    DEFAULT_UI_CONFIG_FILE   = `config.yml`
)

type Color struct {
    Red   float64     `yaml:"red"`
    Green float64     `yaml:"green"`
    Blue  float64     `yaml:"blue"`
    Alpha float64     `yaml:"alpha"`
}

type WindowConfig struct {
    Width       int     `yaml:"width"`
    Height      int     `yaml:"height"`
    X           int     `yaml:"x"`
    Y           int     `yaml:"y"`
    Background  Color   `yaml:"background"`
    Frame       bool    `yaml:"frame"`
    Position    string  `yaml:"position"`
    Resizable   bool    `yaml:"resizable"`
    Stacking    string  `yaml:"stacking"`
    Transparent bool    `yaml:"transparent"`
    Type        string  `yaml:"type"`
}

type Config struct {
    Window WindowConfig `yaml:"window"`
}

var useAlpha    bool
var config      Config = Config{
    Window: WindowConfig{
        Width:        0,
        Height:       0,
        X:           -1,
        Y:           -1,
        Background:  Color{
            Red:   1.0,
            Green: 1.0,
            Blue:  1.0,
            Alpha: 0.0,
        },
        Frame:       true,
        Position:    ``,
        Resizable:   true,
        Stacking:    ``,
        Transparent: false,
        Type:        ``,
    },
}

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

        dc              := diecast.NewServer()
        dc.Address       = c.String(`address`)
        dc.TemplatePath  = c.String(`template-dir`)
        dc.StaticPath    = c.String(`static-dir`)
        dc.ConfigPath    = c.String(`config`)

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

        var window *gtk.Window
        var layout *gtk.Layout

        if obj, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL); err == nil {
            window = obj
        }else{
            log.Fatalf("Failed to create GTK window")
            return
        }

        if obj, err := gtk.LayoutNew(nil, nil); err == nil {
            layout = obj
        }else{
            log.Fatalf("Failed to create GTK layout")
            return
        }

        webview := webkit2.NewWebView()
        webset  := webview.Settings()

        window.Connect(`destroy`, func(){
            gtk.MainQuit()
        })

        window.Connect(`realize`, func(){
        //  move window
            if pos := config.Window.Position; pos == `` {
                if x := config.Window.X; x >= 0 {
                    if y := config.Window.Y; y >= 0 {
                        window.Move(x, y)
                    }
                }
            }else{
                var position gtk.WindowPosition

                switch pos {
                case `center`:
                    position = gtk.WIN_POS_CENTER
                case `mouse`:
                    position = gtk.WIN_POS_MOUSE
                case `center_always`:
                    position = gtk.WIN_POS_CENTER_ALWAYS
                case `center_parent`:
                    position = gtk.WIN_POS_CENTER_ON_PARENT
                default:
                    position = gtk.WIN_POS_NONE
                }

                window.SetPosition(position)
            }

        })

        // window.Connect(`screen-changed`,  UpdateScreen)
        // window.Connect(`expose-event`,    ExposeEvent)


    //  hook up the drawing routines if we're going to be transparent
        if gdkScreen, err := window.GetScreen(); err == nil && gdkScreen.IsComposited() {
            if config.Window.Transparent {
                useAlpha = true

                layout.Connect(`screen-changed`,  OnUpdateScreen)
                layout.Connect(`draw`,            OnDraw)

                webview.Connect(`screen-changed`, OnUpdateScreen)
                webview.Connect(`draw`,           OnDraw)

                window.Connect(`screen-changed`,  OnUpdateScreen)
                window.Connect(`draw`,            OnDraw)

                webview.SetBackgroundColor(gdk.NewRGBA(1.0, 1.0, 1.0, 0.0))
            }
        }else{
            log.Warnf("Failed to get GDK window")
        }

        webset.Set("auto-load-images",                  true)
        webset.Set("enable-plugins",                    true)
        webset.Set("enable-webgl",                      true)
        webset.Set("enable-webaudio",                   true)


        // if webview.GetTransparent() {
        //     log.Debugf("WebKit transparent window enabled")
        // }

        // layout.Add(webview)
        window.Add(webview)

        if len(c.Args()) > 0 {
            webview.LoadURI(c.Args()[0])
        }else{
            webview.LoadURI(fmt.Sprintf("http://%s:%d", dc.Address, dc.Port))
        }

    //  size window
        if w := config.Window.Width; w > 0 {
            if h := config.Window.Height; h > 0 {
                window.SetSizeRequest(w, h)
            }
        }

    //  set window stack preference
        if stacking := config.Window.Stacking; stacking != `` {
            switch stacking {
            case `modal`:
                window.SetModal(true)
            case `above`:
                window.SetKeepAbove(true)
            case `below`:
                window.SetKeepBelow(true)
            }
        }

    //  resizable?
        window.SetResizable(config.Window.Resizable)

    //  decorated? (has window frame)
        window.SetDecorated(config.Window.Frame)

    //  set window type hint
        // if typ := config.Window.Type; typ != `` {
        //     switch typ {
        //     case `normal`:
        //         window.SetTypeHint(gdk.WINDOW_TYPE_HINT_NORMAL)
        //     case `dialog`:
        //         window.SetTypeHint(gdk.WINDOW_TYPE_HINT_DIALOG)
        //     case `menu`:
        //         window.SetTypeHint(gdk.WINDOW_TYPE_HINT_MENU)
        //     case `toolbar`:
        //         window.SetTypeHint(gdk.WINDOW_TYPE_HINT_TOOLBAR)
        //     case `splashscreen`:
        //         window.SetTypeHint(gdk.WINDOW_TYPE_HINT_SPLASHSCREEN)
        //     case `utility`:
        //         window.SetTypeHint(gdk.WINDOW_TYPE_HINT_UTILITY)
        //     case `dock`:
        //         window.SetTypeHint(gdk.WINDOW_TYPE_HINT_DOCK)
        //     case `desktop`:
        //         window.SetTypeHint(gdk.WINDOW_TYPE_HINT_DESKTOP)
        //     case `dropdown_menu`:
        //         window.SetTypeHint(gdk.WINDOW_TYPE_HINT_DROPDOWN_MENU)
        //     case `popup_menu`:
        //         window.SetTypeHint(gdk.WINDOW_TYPE_HINT_POPUP_MENU)
        //     case `tooltip`:
        //         window.SetTypeHint(gdk.WINDOW_TYPE_HINT_TOOLTIP)
        //     case `notification`:
        //         window.SetTypeHint(gdk.WINDOW_TYPE_HINT_NOTIFICATION)
        //     case `combo`:
        //         window.SetTypeHint(gdk.WINDOW_TYPE_HINT_COMBO)
        //     case `dnd`:
        //         window.SetTypeHint(gdk.WINDOW_TYPE_HINT_DND)
        //     }
        // }

        // UpdateWidgetScreen(gtk.WidgetFromNative(unsafe.Pointer(window.ToNative())))
        OnUpdateScreen(window.InitiallyUnowned.Object)

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
            Name:   `config, c`,
            Usage:  `The path to the configuration file`,
            Value:  DEFAULT_UI_CONFIG_FILE,
        },
    }

    app.Run(os.Args)
}


func OnUpdateScreen(widgetObj *glib.Object) {
    if widgetObj != nil {
        widget := gtk.Widget{
            glib.InitiallyUnowned{ widgetObj },
        }

        if screen, err := widget.GetScreen(); err == nil {
            if visual, err := screen.GetRGBAVisual(); err == nil && config.Window.Transparent {
                widget.SetVisual(visual)
                widget.SetAppPaintable(true)

                log.Infof("Alpha visual is available")
                useAlpha = true
            }else{
                log.Warnf("Alpha visual not available")
                useAlpha = false
                return
            }
        }
    }
}

func OnDraw(_ interface{}, context *cairo.Context) {
    if context != nil {
        if useAlpha {
            log.Debugf("Set RGBA: %+v", context)
            context.SetSourceRGBA(config.Window.Background.Red, config.Window.Background.Green, config.Window.Background.Blue, config.Window.Background.Alpha)
        }else{
            context.SetSourceRGB(config.Window.Background.Red, config.Window.Background.Green, config.Window.Background.Blue)
        }

        context.SetOperator(cairo.OPERATOR_SOURCE)
        context.Paint()
    }
}

    // if tgt := ctx.Target(); tgt != nil {
    //     switch tgt.(type) {
    //     case *glib.GObject:
    //         widget := gtk.WidgetFromObject(tgt.(*glib.GObject))
    //         // log.Debugf("expose-event widget: %T %+v", widget, widget)

    //         if gdkWindow := widget.GetWindow(); gdkWindow != nil {
    //             if drawable := gdkWindow.GetDrawable(); drawable != nil {
    //                 context := drawable.CairoCreate()
    //                 // target  := cairo.GetTarget(context)
    //                 target  := gdkWindow.CairoCreateSimilarSurface(cairo.CONTENT_COLOR_ALPHA, gdkWindow.GetWidth(), gdkWindow.GetHeight())

    //                 if surface := cairo.NewSurfaceFromC(target, context); surface != nil {
    //                     if useAlpha {
    //                         surface.SetSourceRGBA(config.Window.Background.Red, config.Window.Background.Green, config.Window.Background.Blue, config.Window.Background.Alpha)
    //                     }else{
    //                         surface.SetSourceRGB(config.Window.Background.Red, config.Window.Background.Green, config.Window.Background.Blue)
    //                     }

    //                     surface.SetOperator(cairo.OPERATOR_SOURCE)
    //                     surface.Paint()
    //                     cairo.Destroy(context)
    //                 }else{
    //                     log.Debugf("expose-event: failed to create cairo surface of %+v", widget)
    //                 }
    //             }else{
    //                 log.Debugf("expose-event: failed to get drawable surface of %+v", widget)
    //             }
    //         }else{
    //             log.Debugf("expose-event: failed to get parent window of %+v", widget)
    //         }
    //     default:
    //         log.Debugf("expose-event: expected *glib.GObject target, got %T", tgt)
    //     }
    // }