package main

import (
    "fmt"
    "io/ioutil"
    "net"
    "os"
    "strconv"
    "strings"
    "unsafe"

    "github.com/auroralaboratories/corona-ui/util"
    "github.com/auroralaboratories/go-cairo"
    "github.com/auroralaboratories/go-gtk/gdk"
    "github.com/auroralaboratories/go-gtk/glib"
    "github.com/auroralaboratories/go-gtk/gtk"
    "github.com/auroralaboratories/go-webkit/webkit"
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

        window  := gtk.NewWindow(gtk.WINDOW_TOPLEVEL)
        layout  := gtk.NewAlignment(1.0, 1.0, 1.0, 1.0)

        webview := webkit.NewWebView()
        webset  := webkit.NewWebSettings()

        window.Connect(`destroy`,         gtk.MainQuit)
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
        if config.Window.Transparent {
            layout.Connect(`screen-changed`,  UpdateScreen)
            layout.Connect(`expose-event`,    ExposeEvent)

            webview.Connect(`screen-changed`, UpdateScreen)
            webview.Connect(`expose-event`,   ExposeEvent)

            layout.SetAppPaintable(true)
            window.SetAppPaintable(true)
        }

        webview.Connect(`resource-load-finished`, func() {
            log.Infof("Loaded %s", webview.GetUri())
        })

        webset.Set("auto-load-images",                  true)
        webset.Set("auto-resize-window",                false)
        webset.Set("enable-plugins",                    true)
        webset.Set("enable-scripts",                    true)
        webset.Set("enable-accelerated-compositing",    false)
        webset.Set("enable-webgl",                      true)
        webset.Set("enable-webaudio",                   true)
        webset.Set("enable-file-access-from-file-uris", true)

        webview.SetSettings(webset)
        webview.SetTransparent(config.Window.Transparent)

        if webview.GetTransparent() {
            log.Debugf("WebKit transparent window enabled")
        }

        layout.Add(webview)
        window.Add(layout)

        if len(c.Args()) > 0 {
            webview.LoadUri(c.Args()[0])
        }else{
            webview.LoadUri(fmt.Sprintf("http://%s:%d", dc.Address, dc.Port))
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
        if typ := config.Window.Type; typ != `` {
            switch typ {
            case `normal`:
                window.SetTypeHint(gdk.WINDOW_TYPE_HINT_NORMAL)
            case `dialog`:
                window.SetTypeHint(gdk.WINDOW_TYPE_HINT_DIALOG)
            case `menu`:
                window.SetTypeHint(gdk.WINDOW_TYPE_HINT_MENU)
            case `toolbar`:
                window.SetTypeHint(gdk.WINDOW_TYPE_HINT_TOOLBAR)
            case `splashscreen`:
                window.SetTypeHint(gdk.WINDOW_TYPE_HINT_SPLASHSCREEN)
            case `utility`:
                window.SetTypeHint(gdk.WINDOW_TYPE_HINT_UTILITY)
            case `dock`:
                window.SetTypeHint(gdk.WINDOW_TYPE_HINT_DOCK)
            case `desktop`:
                window.SetTypeHint(gdk.WINDOW_TYPE_HINT_DESKTOP)
            case `dropdown_menu`:
                window.SetTypeHint(gdk.WINDOW_TYPE_HINT_DROPDOWN_MENU)
            case `popup_menu`:
                window.SetTypeHint(gdk.WINDOW_TYPE_HINT_POPUP_MENU)
            case `tooltip`:
                window.SetTypeHint(gdk.WINDOW_TYPE_HINT_TOOLTIP)
            case `notification`:
                window.SetTypeHint(gdk.WINDOW_TYPE_HINT_NOTIFICATION)
            case `combo`:
                window.SetTypeHint(gdk.WINDOW_TYPE_HINT_COMBO)
            case `dnd`:
                window.SetTypeHint(gdk.WINDOW_TYPE_HINT_DND)
            }
        }

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
            Name:   `config, c`,
            Usage:  `The path to the configuration file`,
            Value:  DEFAULT_UI_CONFIG_FILE,
        },
    }

    app.Run(os.Args)
}


func UpdateScreen(ctx *glib.CallbackContext) {
    if tgt := ctx.Target(); tgt != nil {
        widget := gtk.WidgetFromObject(tgt.(*glib.GObject))
        UpdateWidgetScreen(widget)
    }else{
        log.Debugf("screen-changed fired without a target")
    }
}

func UpdateWidgetScreen(widget *gtk.Widget) {
    screen := widget.GetScreen()
    var colormap *gdk.Colormap

    if screen.IsComposited() && config.Window.Transparent {
        log.Infof("Compositing is enabled")
        useAlpha = true
        colormap = screen.GetRGBAColormap()
    }else{
        log.Warnf("Compositing is disabled")
        useAlpha = false
        colormap = screen.GetRGBColormap()
    }

    widget.SetColormap(colormap)
}

func ExposeEvent(ctx *glib.CallbackContext) {
    if tgt := ctx.Target(); tgt != nil {
        switch tgt.(type) {
        case *glib.GObject:
            widget := gtk.WidgetFromObject(tgt.(*glib.GObject))
            // log.Debugf("expose-event widget: %T %+v", widget, widget)

            if gdkWindow := widget.GetWindow(); gdkWindow != nil {
                if drawable := gdkWindow.GetDrawable(); drawable != nil {
                    context := drawable.CairoCreate()
                    // target  := cairo.GetTarget(context)
                    target  := gdkWindow.CairoCreateSimilarSurface(cairo.CONTENT_COLOR_ALPHA, gdkWindow.GetWidth(), gdkWindow.GetHeight())

                    if surface := cairo.NewSurfaceFromC(target, context); surface != nil {
                        if useAlpha {
                            surface.SetSourceRGBA(config.Window.Background.Red, config.Window.Background.Green, config.Window.Background.Blue, config.Window.Background.Alpha)
                        }else{
                            surface.SetSourceRGB(config.Window.Background.Red, config.Window.Background.Green, config.Window.Background.Blue)
                        }

                        surface.SetOperator(cairo.OPERATOR_SOURCE)
                        surface.Paint()
                        cairo.Destroy(context)
                    }else{
                        log.Debugf("expose-event: failed to create cairo surface of %+v", widget)
                    }
                }else{
                    log.Debugf("expose-event: failed to get drawable surface of %+v", widget)
                }
            }else{
                log.Debugf("expose-event: failed to get parent window of %+v", widget)
            }
        default:
            log.Debugf("expose-event: expected *glib.GObject target, got %T", tgt)
        }
    }
}