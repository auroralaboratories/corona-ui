package main

import (
    "fmt"
    "image"

    "github.com/auroralaboratories/go-webkit2/webkit2"
    "github.com/gotk3/gotk3/cairo"
    "github.com/gotk3/gotk3/gdk"
    "github.com/gotk3/gotk3/glib"
    "github.com/gotk3/gotk3/gtk"
    log "github.com/Sirupsen/logrus"
)

type Window struct {
    Config              *WindowConfig
    ShapeAlphaThreshold uint8
    URI                 string
    Server              *Server

    gtkWindow *gtk.Window
    layout    *gtk.Layout
    webview   *webkit2.WebView
    webset    *webkit2.Settings
}

func NewWindow(server *Server) *Window {
    return &Window{
        Server: server,
    }
}

func (self *Window) Initialize(config *WindowConfig) error {
    self.Config = config

    gtk.Init(nil)

    if obj, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL); err == nil {
        self.gtkWindow = obj
    }else{
        return fmt.Errorf("Failed to create GTK window")
    }

    if obj, err := gtk.LayoutNew(nil, nil); err == nil {
        self.layout = obj
    }else{
        return fmt.Errorf("Failed to create GTK layout")
    }

    self.webview = webkit2.NewWebView()
    self.webset  = self.webview.Settings()

    self.gtkWindow.Connect(`destroy`, func(){
        gtk.MainQuit()
    })

    self.gtkWindow.Connect(`realize`, func(){
    //  move window
        if pos := self.Config.Position; pos == `` {
            if x := self.Config.X; x >= 0 {
                if y := self.Config.Y; y >= 0 {
                    self.gtkWindow.Move(x, y)
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

            self.gtkWindow.SetPosition(position)
        }

    })

//  hook up the drawing routines if we're going to be transparent
    if gdkScreen, err := self.gtkWindow.GetScreen(); err == nil && gdkScreen.IsComposited() {
        if self.Config.Transparent {
            useAlpha = false

            self.layout.Connect(`screen-changed`,     self.onUpdateScreen)
            self.layout.Connect(`draw`,               self.onDraw)

            self.webview.Connect(`screen-changed`,    self.onUpdateScreen)
            self.webview.Connect(`draw`,              self.onDraw)

            self.gtkWindow.Connect(`screen-changed`,  self.onUpdateScreen)
            self.gtkWindow.Connect(`draw`,            self.onDraw)

            self.webview.SetBackgroundColor(gdk.NewRGBA(1.0, 1.0, 1.0, 0.0))
        }
    }else{
        log.Warnf("Failed to get GDK window")
    }

    self.webset.Set("auto-load-images",                  true)
    self.webset.Set("enable-plugins",                    true)
    self.webset.Set("enable-webgl",                      true)
    self.webset.Set("enable-webaudio",                   true)


    // if webview.GetTransparent() {
    //     log.Debugf("WebKit transparent window enabled")
    // }

    // layout.Add(webview)
    self.gtkWindow.Add(self.webview)

    if self.URI != `` {
        self.webview.LoadURI(self.URI)
    }else if self.Server != nil {
        self.webview.LoadURI(fmt.Sprintf("http://%s:%d", self.Server.Address, self.Server.Port))
    }

//  size window
    if w := self.Config.Width; w > 0 {
        if h := self.Config.Height; h > 0 {
            self.gtkWindow.SetSizeRequest(w, h)
        }
    }

//  set window stack preference
    if stacking := self.Config.Stacking; stacking != `` {
        switch stacking {
        case `modal`:
            self.gtkWindow.SetModal(true)
        case `above`:
            self.gtkWindow.SetKeepAbove(true)
        case `below`:
            self.gtkWindow.SetKeepBelow(true)
        }
    }

//  resizable?
    self.gtkWindow.SetResizable(self.Config.Resizable)

//  decorated? (has window frame)
    self.gtkWindow.SetDecorated(self.Config.Frame)

//  set window type hint
    // if typ := self.Config.Type; typ != `` {
    //     switch typ {
    //     case `normal`:
    //         self.gtkWindow.SetTypeHint(gdk.WINDOW_TYPE_HINT_NORMAL)
    //     case `dialog`:
    //         self.gtkWindow.SetTypeHint(gdk.WINDOW_TYPE_HINT_DIALOG)
    //     case `menu`:
    //         self.gtkWindow.SetTypeHint(gdk.WINDOW_TYPE_HINT_MENU)
    //     case `toolbar`:
    //         self.gtkWindow.SetTypeHint(gdk.WINDOW_TYPE_HINT_TOOLBAR)
    //     case `splashscreen`:
    //         self.gtkWindow.SetTypeHint(gdk.WINDOW_TYPE_HINT_SPLASHSCREEN)
    //     case `utility`:
    //         self.gtkWindow.SetTypeHint(gdk.WINDOW_TYPE_HINT_UTILITY)
    //     case `dock`:
    //         self.gtkWindow.SetTypeHint(gdk.WINDOW_TYPE_HINT_DOCK)
    //     case `desktop`:
    //         self.gtkWindow.SetTypeHint(gdk.WINDOW_TYPE_HINT_DESKTOP)
    //     case `dropdown_menu`:
    //         self.gtkWindow.SetTypeHint(gdk.WINDOW_TYPE_HINT_DROPDOWN_MENU)
    //     case `popup_menu`:
    //         self.gtkWindow.SetTypeHint(gdk.WINDOW_TYPE_HINT_POPUP_MENU)
    //     case `tooltip`:
    //         self.gtkWindow.SetTypeHint(gdk.WINDOW_TYPE_HINT_TOOLTIP)
    //     case `notification`:
    //         self.gtkWindow.SetTypeHint(gdk.WINDOW_TYPE_HINT_NOTIFICATION)
    //     case `combo`:
    //         self.gtkWindow.SetTypeHint(gdk.WINDOW_TYPE_HINT_COMBO)
    //     case `dnd`:
    //         self.gtkWindow.SetTypeHint(gdk.WINDOW_TYPE_HINT_DND)
    //     }
    // }

    self.onUpdateScreen(self.gtkWindow.InitiallyUnowned.Object)
    return nil
}

func (self *Window) Show() error {
    self.gtkWindow.ShowAll()

    // proxy := os.Getenv(`HTTP_PROXY`)
    // if len(proxy) > 0 {
    //     soup_uri := webkit.SoupUri(proxy)
    //     webkit.GetDefaultSession().Set(`proxy-uri`, soup_uri)
    //     soup_uri.Free()
    // }

    gtk.Main()
    return nil
}

func (self *Window) onUpdateScreen(widgetObj *glib.Object) {
    if widgetObj != nil {
        widget := gtk.Widget{
            glib.InitiallyUnowned{ widgetObj },
        }

        if screen, err := widget.GetScreen(); err == nil {
            if visual, err := screen.GetRGBAVisual(); err == nil && self.Config.Transparent {
                widget.SetVisual(visual)
                widget.SetAppPaintable(true)

                log.Debugf("Alpha visual is available")
                useAlpha = true
            }else{
                log.Debugf("Alpha visual not available")
                useAlpha = false
                return
            }
        }
    }
}

func (self *Window) onDraw(_ interface{}, context *cairo.Context) {
    if context != nil {
        if useAlpha {
            context.SetSourceRGBA(self.Config.Background.Red, self.Config.Background.Green, self.Config.Background.Blue, self.Config.Background.Alpha)
        }else{
            context.SetSourceRGB(self.Config.Background.Red, self.Config.Background.Green, self.Config.Background.Blue)
        }

        context.SetOperator(cairo.OPERATOR_SOURCE)
        context.Paint()
    }
}


func (self *Window) updateWindowShapePixmapFromWebview() error {
    self.webview.GetSnapshot(func(result *image.RGBA, err error){
        if err == nil {
            j := 0
            shapemask := make([]uint8, len(result.Pix)/4)

        //  image's Pix[] is a slice of R,G,B,A values; we're only looking for alpha
        //  so start at Pix[3] and step forward by 4 elements each time
            for i := 3; i < len(result.Pix); i=i+4 {
                if result.Pix[i] <= self.ShapeAlphaThreshold {
                    // shapemask[j] set to off
                    shapemask[j] = 0
                }else{
                    shapemask[j] = 1
                }

                j += 1
            }
        }
    })

    return nil
}