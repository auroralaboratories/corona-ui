package main

import (
    "fmt"
    "image"
    "math"
    "time"
    // "io/ioutil"
    "io"
    "os"

    "github.com/auroralaboratories/go-webkit2/webkit2"
    "github.com/BurntSushi/xgb/xproto"
    "github.com/BurntSushi/xgb/shape"
    "github.com/BurntSushi/xgbutil"
    "github.com/BurntSushi/xgbutil/xgraphics"
    "github.com/gotk3/gotk3/cairo"
    "github.com/gotk3/gotk3/gdk"
    "github.com/gotk3/gotk3/glib"
    "github.com/gotk3/gotk3/gtk"
    log "github.com/Sirupsen/logrus"
)

type Color struct {
    Red   float64     `yaml:"red"`
    Green float64     `yaml:"green"`
    Blue  float64     `yaml:"blue"`
    Alpha float64     `yaml:"alpha"`
}

type WindowConfig struct {
    Width          int     `yaml:"width"`
    Height         int     `yaml:"height"`
    X              int     `yaml:"x"`
    Y              int     `yaml:"y"`
    Background     Color   `yaml:"background"`
    Frame          bool    `yaml:"frame"`
    Position       string  `yaml:"position"`
    Resizable      bool    `yaml:"resizable"`
    Stacking       string  `yaml:"stacking"`
    Transparent    bool    `yaml:"transparent"`
    Shaped         bool    `yaml:"shaped"`
    KnockoutLimit  uint8   `yaml:"knockout_limit"`
    Type           string  `yaml:"type"`
}

type Window struct {
    Config              *WindowConfig
    URI                 string
    Server              *Server

    xconn     *xgbutil.XUtil
    gtkWindow *gtk.Window
    layout    *gtk.Layout
    webview   *webkit2.WebView
    webset    *webkit2.Settings

    shapeOk   bool
    winShape  xproto.Pixmap
}

func NewWindow(server *Server) *Window {
    return &Window{
        Server:         server,
    }
}

func (self *Window) Initialize(config *WindowConfig) error {
    self.Config = config

    if xconn, err := xgbutil.NewConn(); err == nil {
        self.xconn = xconn

        if err := shape.Init(self.xconn.Conn()); err == nil {
            self.shapeOk = true
        }else{
            self.shapeOk = false
            self.Config.Shaped = false
            log.Warnf("Failed to initialize X11 SHAPE module, cannot change window shape: %v", err)
        }
    }else{
        return fmt.Errorf("Failed to connect to X11: %v", err)
    }

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

    self.webview.Connect(`load-changed`, func(_ *glib.Object, event webkit2.LoadEvent){
        log.Debugf("Load event %d", event)

        switch event {
        case webkit2.LoadFinished:
            go func(){
                for {
                    self.updateWindowShapePixmapFromWebview()
                    time.Sleep(3 * time.Second)
                }
            }()
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
    self.webview.GetSnapshotWithOptions(webkit2.RegionFullDocument, webkit2.SnapshotOptionTransparentBackground, func(result *image.RGBA, err error){
        if err == nil {
        //  so...
        //  xgbutil doesn't support creating 1-bit depth pixmaps (bitmaps)
        //  This means we need to do this manually...                        :(
            if pixmapId, err := xproto.NewPixmapId(self.xconn.Conn()); err == nil {
                width    := result.Bounds().Dx()
                height   := result.Bounds().Dy()
                drawable := xproto.Drawable(self.xconn.RootWin())

            //  synchronously create the pixmap with the appropriate dimensions and color depth
                if err := xproto.CreatePixmapChecked(self.xconn.Conn(),
                    1,                                                  // color depth (bits)
                    pixmapId,                                           // allocated pixmap ID
                    drawable,                                           // pixmap drawable
                    uint16(width),
                    uint16(height)).Check(); err == nil {
                    bitmap     := make([]byte, (width*height)/8)
                    bitmapBits := 0
                    bitmapI    := 0
                    ko_Limit   := self.Config.KnockoutLimit

                    c_off      := 0
                    c_on       := 0

                //  image's Pix[] is a slice of R,G,B,A values; we're only looking for alpha
                //  so start at Pix[3] and step forward by 4 elements each time
                //  populate the bitmap to represent either part-of or not-part-of the window area
                    for i := 3; i < len(result.Pix); i=i+4 {
                        currentBit := byte(math.Mod(float64(bitmapBits), 8.0))

                    //  if the current result alpha value is gt the knockout threshold, flag the
                    //  corresponding bitmap bit ON
                        if result.Pix[i] <= ko_Limit {
                            c_off += 1
                        }else{
                            bitmap[bitmapI] |= (byte(1)<<currentBit)
                            c_on += 1
                        }

                        bitmapBits += 1

                    //  increment the current bitmap index every 8 bits processed
                        if currentBit == 7 {
                            bitmapI += 1
                        }
                    }

                    if file, err := os.OpenFile(`./bitmap.xbm`, os.O_WRONLY, 0755); err == nil {
                        io.WriteString(file, fmt.Sprintf("#define max_width %d\n", width))
                        io.WriteString(file, fmt.Sprintf("#define max_height %d\n", height))
                        io.WriteString(file, "static unsigned char max_bits[] = {\n")

                        for i := 0; i < len(bitmap); i++ {
                            io.WriteString(file, fmt.Sprintf("0x%x", bitmap[i]))

                            if (i+1) < len(bitmap) {
                                io.WriteString(file, ", ")
                            }else{
                                io.WriteString(file, "};\n")
                            }
                        }

                        file.Close()
                    }else{
                        log.Warnf("Unable to create XBM: %v", err)
                    }


                    log.Debugf("Resulting bitmap: %d bytes (%d bits); %d opaque, %d transparent", len(bitmap), len(bitmap)*8, c_on, c_off)

                //  resulting bitmap should be exactly 1/32 the size of the incoming image
                //  (32-bit RGBA is 32x the size of a 1-bit bitmap)
                //
                    if len(bitmap) == (len(result.Pix)/32) {
                    //  this is a reimplementation of xgraphics.XDraw() because I'm not ready to fork
                    //  xgbutil yet
                    //
                        var chunk []byte

                        rowsPerPut   := ((xgbutil.MaxReqSize - 28) / width)
                        bytesPerPut  := (rowsPerPut * width)
                        xpos         := result.Rect.Min.X
                        ypos         := result.Rect.Min.Y
                        heightPerPut := 0
                        start        := 0
                        end          := 0

                        for len(bitmap) > end {
                        //  end is offset by current position + chunk size
                            end = start + bytesPerPut

                        //  make sure end doesn't extend beyond data
                            if len(bitmap) < end {
                                end = len(bitmap)
                            }

                            chunk        = bitmap[start:end]
                            heightPerPut = (len(chunk) / width)

                        //  put bitmap data into the destination drawable
                        //  this is the nutmeat of an unchecked XDraw request
                            xproto.PutImage(self.xconn.Conn(),
                                xproto.ImageFormatXYBitmap,                 // format
                                xproto.Drawable(pixmapId),                  // drawable surface
                                self.xconn.GC(),                            // graphics context
                                uint16(width),                              // width
                                uint16(heightPerPut),                       // height of this chunk
                                int16(xpos),                                // origin X of this chunk
                                int16(ypos),                                // origin Y of this chunk
                                0,                                          // XGB: LeftPad
                                1,                                          // color depth (bits)
                                chunk)

                        //  new start is this end
                            start = end

                        //  next row (y-position) increments by however many rows we just put
                            ypos += rowsPerPut
                        }


                        if ximage, err := xgraphics.NewDrawable(self.xconn, xproto.Drawable(pixmapId)); err == nil {
                            log.Debugf("DUMP shape to ./test.png")
                            ximage.SavePng(`./test.png`)
                        }else{
                            log.Warnf("Failed to dump image: %v", err)
                        }

                        if err := self.applyWindowShape(pixmapId); err != nil {
                            log.Errorf("Failed to apply window shape: %v", err)
                        }
                    }else{
                        log.Errorf("Failed to create X11 pixmap: resulting bitmap is the wrong size (expected %d, got %d)", (len(result.Pix)/4), len(bitmap))
                    }
                }else{
                    log.Errorf("Failed to create X11 pixmap: %v", err)
                }
            }else{
                log.Errorf("Failed to allocate X11 pixmap ID: %v", err)
            }
        }
    })

    return nil
}

func (self *Window) applyWindowShape(pixmap xproto.Pixmap) error {
    if self.shapeOk {
        if pixmap > 0 {
            defer xgraphics.FreePixmap(self.xconn, pixmap)

            if gdkWindow, err := self.gtkWindow.GetWindow(); err == nil {
                if xid, err := gdkWindow.GetWindowID(); err == nil {
                //  shape.Op(0)   -> ShapeSet
                //  shape.Kind(0) -> ShapeBounding
                //
                    log.Debugf("Shape pixmap: %v", pixmap)

                    if err := shape.MaskChecked(self.xconn.Conn(), shape.Op(0), shape.Kind(0), xproto.Window(xid), 0, 0, pixmap).Check(); err != nil {
                        return err
                    }

                }else{
                    return err
                }
            }else{
                return err
            }
        }else{
            return fmt.Errorf("Window shape pixmap is unset")
        }
    }else{
        return fmt.Errorf("Cannot apply shape; SHAPE extension not initialized")
    }

    return nil
}