package main

import (
	"fmt"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/auroralaboratories/go-webkit2/webkit2"
	"github.com/auroralaboratories/gotk3/cairo"
	"github.com/auroralaboratories/gotk3/gdk"
	"github.com/auroralaboratories/gotk3/glib"
	"github.com/auroralaboratories/gotk3/gtk"
	"strconv"
	"strings"
)

type Color struct {
	Red   float64 `json:"red"`
	Green float64 `json:"green"`
	Blue  float64 `json:"blue"`
	Alpha float64 `json:"alpha"`
}

type Reservation struct {
	Left   string `json:"left"`
	Right  string `json:"right"`
	Top    string `json:"top"`
	Bottom string `json:"bottom"`
}

type Monitor struct {
	Number  int  `json:"number"`
	Primary bool `json:"primary"`
	Width   uint `json:"width"`
	Height  uint `json:"height"`
	X       int  `json:"x"`
	Y       int  `json:"y"`
}

func (self Monitor) String() string {
	pri := ``

	if self.Primary {
		pri = ` (primary)`
	}

	return fmt.Sprintf(
		"Monitor %d: %dx%d+%d+%d%s",
		self.Number,
		self.Width,
		self.Height,
		self.X,
		self.Y,
		pri,
	)
}

type WindowConfig struct {
	Width         string       `json:"width,omitempty"`
	Height        string       `json:"height,omitempty"`
	X             string       `json:"x,omitempty"`
	Y             string       `json:"y,omitempty"`
	Monitor       int          `json:"monitor"`
	Background    *Color       `json:"background,omitempty"`
	Frame         bool         `json:"frame"`
	Position      string       `json:"position,omitempty"`
	Resizable     bool         `json:"resizable"`
	Stacking      string       `json:"stacking,omitempty"`
	Transparent   bool         `json:"transparent"`
	Shaped        bool         `json:"shaped"`
	Reserve       bool         `json:"reserve"`
	ReserveBounds *Reservation `json:"bounds,omitempty"`
	Type          string       `json:"type,omitempty"`
}

type Window struct {
	Config      *WindowConfig `json:"config"`
	URI         string        `json:"uri"`
	Server      *Server       `json:"server"`
	Realized    bool          `json:"realized"`
	Width       int           `json:"width"`
	Height      int           `json:"height"`
	X           int           `json:"x"`
	Y           int           `json:"y"`
	Monitors    []Monitor     `json:"monitors"`
	ScreenWidth int           `json:"screen_width"`
	ScreeHeight int           `json:"screen_height"`
	xconn       *xgbutil.XUtil
	gtkWindow   *gtk.Window
	layout      *gtk.Layout
	webview     *webkit2.WebView
	webset      *webkit2.Settings
}

func NewWindow(server *Server) *Window {
	return &Window{
		Server:   server,
		Monitors: make([]Monitor, 0),
	}
}

func (self *Window) Initialize(config *WindowConfig) error {
	self.Config = config

	if xconn, err := xgbutil.NewConn(); err == nil {
		self.xconn = xconn
	} else {
		return fmt.Errorf("Failed to connect to X11: %v", err)
	}

	gtk.Init(nil)

	if obj, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL); err == nil {
		self.gtkWindow = obj
	} else {
		return fmt.Errorf("Failed to create GTK window")
	}

	if obj, err := gtk.LayoutNew(nil, nil); err == nil {
		self.layout = obj
	} else {
		return fmt.Errorf("Failed to create GTK layout")
	}

	self.webview = webkit2.NewWebView()
	self.webset = self.webview.Settings()

	self.gtkWindow.Connect(`destroy`, func() {
		gtk.MainQuit()
	})

	self.gtkWindow.Connect(`configure-event`, func() {
		self.onResizeOrMove()
	})

	self.gtkWindow.Connect(`realize`, func() {
		if gdkScreen, err := self.gtkWindow.GetScreen(); err == nil {
			numMonitors := gdkScreen.GetNMonitors()
			primaryMonitor := gdkScreen.GetPrimaryMonitor()

			for i := 0; i < numMonitors; i++ {
				if geom := gdkScreen.GetMonitorGeometry(i); geom != nil {
					monitor := Monitor{
						Number:  i,
						Primary: (i == primaryMonitor),
						Width:   uint(geom.GetWidth()),
						Height:  uint(geom.GetHeight()),
						X:       geom.GetX(),
						Y:       geom.GetY(),
					}

					log.Debugf("%v", monitor)
					self.Monitors = append(self.Monitors, monitor)

					if (self.Config.Monitor < 0 && i == primaryMonitor) || (i == self.Config.Monitor) {
						self.ScreenWidth = geom.GetWidth()
						self.ScreeHeight = geom.GetHeight()
					}
				}
			}

			self.Realized = true

			self.onResizeOrMove()
		} else {
			log.Error(err)
		}

		//  move/resize window
		if pos := self.Config.Position; pos == `` {
			if x, err := self.DimensionToInt(self.Config.X, self.ScreenWidth); err == nil {
				if y, err := self.DimensionToInt(self.Config.Y, self.ScreeHeight); err == nil {
					log.Debugf("Offsetting to monitor %d", self.Config.Monitor)

					if self.Config.Monitor >= 0 && self.Config.Monitor < len(self.Monitors) {
						x += self.Monitors[self.Config.Monitor].X
						y += self.Monitors[self.Config.Monitor].Y
					}

					if x > 0 || y > 0 {
						log.Debugf("Moving to %d,%d", x, y)
						self.gtkWindow.Move(x, y)
					}
				} else {
					log.Error(err)
				}
			} else {
				log.Error(err)
			}
		} else {
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

		//  size window
		if w, err := self.DimensionToInt(self.Config.Width, self.ScreenWidth); err == nil && w > 0 {
			if h, err := self.DimensionToInt(self.Config.Height, self.ScreeHeight); err == nil && h > 0 {
				self.gtkWindow.SetDefaultSize(w, h)
			}
		}

		//  set window struts (reserve)
		if self.Config.Reserve {
			if gdkWindow, err := self.gtkWindow.GetWindow(); err == nil {
				if xid, err := gdkWindow.GetWindowID(); err == nil {
					strutPartial := ewmh.WmStrutPartial{}
					strutSimple := ewmh.WmStrut{}

					if v, err := self.DimensionToInt(self.Config.ReserveBounds.Left, self.ScreenWidth); err == nil {
						strutPartial.Left = uint(v)
					} else {
						log.Warningf("Invalid left reservation value: %v", err)
					}

					if v, err := self.DimensionToInt(self.Config.ReserveBounds.Right, self.ScreenWidth); err == nil {
						strutPartial.Right = uint(v)
					} else {
						log.Warningf("Invalid right reservation value: %v", err)
					}

					if v, err := self.DimensionToInt(self.Config.ReserveBounds.Top, self.ScreeHeight); err == nil {
						strutPartial.Top = uint(v)
					} else {
						log.Warningf("Invalid top reservation value: %v", err)
					}

					if v, err := self.DimensionToInt(self.Config.ReserveBounds.Bottom, self.ScreeHeight); err == nil {
						strutPartial.Bottom = uint(v)
					} else {
						log.Warningf("Invalid bottom reservation value: %v", err)
					}

					strutSimple.Left = strutPartial.Left
					strutSimple.Right = strutPartial.Right
					strutSimple.Top = strutPartial.Top
					strutSimple.Bottom = strutPartial.Bottom

					if strutPartial.Left > 0 {
						strutPartial.LeftStartY = 0
						strutPartial.LeftEndY = uint(self.ScreeHeight)
					}

					if strutPartial.Right > 0 {
						strutPartial.RightStartY = 0
						strutPartial.RightEndY = uint(self.ScreeHeight)
					}

					if strutPartial.Top > 0 {
						strutPartial.TopStartX = 0
						strutPartial.TopEndX = uint(self.ScreenWidth)
					}

					if strutPartial.Bottom > 0 {
						strutPartial.BottomStartX = 0
						strutPartial.BottomEndX = uint(self.ScreenWidth)
					}

					log.Debugf("Setting window struts: %+v / %+v", strutPartial, strutSimple)

					if err := ewmh.WmStrutPartialSet(self.xconn, xproto.Window(xid), &strutPartial); err != nil {
						log.Errorf("Failed to set window space reservations (partials): %v", err)
					}

					if err := ewmh.WmStrutSet(self.xconn, xproto.Window(xid), &strutSimple); err != nil {
						log.Errorf("Failed to set window space reservations: %v", err)
					}

				} else {
					log.Errorf("Failed to retrieve X11 window ID: %v", err)
				}
			} else {
				log.Errorf("Failed to retrieve GDK window: %v", err)
			}
		}
	})

	self.webview.Connect(`load-changed`, func(_ *glib.Object, event webkit2.LoadEvent) {
		log.Debugf("Load event %d", event)

		switch event {
		case webkit2.LoadFinished:
			self.RefreshShape()
		}
	})

	//  hook up the drawing routines if we're going to be transparent
	if gdkScreen, err := self.gtkWindow.GetScreen(); err == nil && gdkScreen.IsComposited() {
		if self.Config.Transparent {
			useAlpha = false

			self.layout.Connect(`screen-changed`, self.onUpdateScreen)
			self.layout.Connect(`draw`, self.onDraw)

			self.webview.Connect(`screen-changed`, self.onUpdateScreen)
			self.webview.Connect(`draw`, self.onDraw)

			self.gtkWindow.Connect(`screen-changed`, self.onUpdateScreen)
			self.gtkWindow.Connect(`draw`, self.onDraw)

			self.webview.SetBackgroundColor(gdk.NewRGBA(1.0, 1.0, 1.0, 0.0))
		}
	} else {
		log.Warningf("Failed to get GDK window")
	}

	self.webset.Set(`auto-load-image`, true)
	self.webset.Set(`enable-javascript`, true)
	self.webset.Set(`enable-plugins`, true)
	self.webset.Set(`enable-webaudio`, true)
	self.webset.Set(`enable-webgl`, true)

	// if webview.GetTransparent() {
	//     log.Debugf("WebKit transparent window enabled")
	// }

	// layout.Add(webview)
	self.gtkWindow.Add(self.webview)

	if self.URI == `` {
		self.URI = fmt.Sprintf("http://%s", self.Server.Address)
	}

	if self.Server != nil {
		self.webview.LoadURI(self.URI)
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

	// set default size
	self.gtkWindow.SetDefaultSize(-1, -1)

	//  set window type hint
	if typ := self.Config.Type; typ != `` {
		switch typ {
		case `normal`:
			self.gtkWindow.SetTypeHint(gdk.TypeHintNormal)
		case `dialog`:
			self.gtkWindow.SetTypeHint(gdk.TypeHintDialog)
		case `menu`:
			self.gtkWindow.SetTypeHint(gdk.TypeHintMenu)
		case `toolbar`:
			self.gtkWindow.SetTypeHint(gdk.TypeHintToolbar)
		case `splashscreen`:
			self.gtkWindow.SetTypeHint(gdk.TypeHintSplashscreen)
		case `utility`:
			self.gtkWindow.SetTypeHint(gdk.TypeHintUtility)
		case `dock`:
			self.gtkWindow.SetTypeHint(gdk.TypeHintDock)
		case `desktop`:
			self.gtkWindow.SetTypeHint(gdk.TypeHintDesktop)
		case `dropdown_menu`:
			self.gtkWindow.SetTypeHint(gdk.TypeHintDropdown_menu)
		case `popup_menu`:
			self.gtkWindow.SetTypeHint(gdk.TypeHintPopup_menu)
		case `tooltip`:
			self.gtkWindow.SetTypeHint(gdk.TypeHintTooltip)
		case `notification`:
			self.gtkWindow.SetTypeHint(gdk.TypeHintNotification)
		case `combo`:
			self.gtkWindow.SetTypeHint(gdk.TypeHintCombo)
		case `dnd`:
			self.gtkWindow.SetTypeHint(gdk.TypeHintDnd)
		}
	}

	self.onUpdateScreen(self.gtkWindow.InitiallyUnowned.Object)
	return nil
}

func (self *Window) DimensionToInt(expr string, relativeTo int) (int, error) {
	if expr == `` {
		expr = `0`
	}

	if v, err := strconv.ParseInt(expr, 10, 64); err == nil {
		log.Debugf("Absolute dimension: %d", v)
		return int(v), nil
	} else {
		if strings.HasSuffix(expr, `%`) {
			if self.Realized {
				if perc, err := strconv.ParseFloat(expr[0:len(expr)-1], 64); err == nil {
					out := ((perc / 100.0) * float64(relativeTo))
					log.Debugf("Relative dimension: %g%% of %d = %g", perc, relativeTo, out)
					return int(out), nil
				} else {
					log.Warningf("%v", err)
					return -1, err
				}
			} else {
				err := fmt.Errorf("Cannot use relative dimensions with unrealized window")
				log.Warningf("%v", err)
				return -1, err
			}
		}

		return -1, err
	}
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

func (self *Window) RefreshShape() {
	if self.Config.Shaped {
		self.updateWindowShapePixmapFromWebview()
	}
}

func (self *Window) onResizeOrMove() {
	if self.Realized {
		self.X, self.Y = self.gtkWindow.GetPosition()
		self.Width, self.Height = self.gtkWindow.GetSize()
	}
}

func (self *Window) onUpdateScreen(widgetObj *glib.Object) {
	if widgetObj != nil {
		widget := gtk.Widget{
			glib.InitiallyUnowned{widgetObj},
		}

		if screen, err := widget.GetScreen(); err == nil {
			if visual, err := screen.GetRGBAVisual(); err == nil && self.Config.Transparent {
				widget.SetVisual(visual)
				widget.SetAppPaintable(true)

				log.Debugf("Alpha visual is available")
				useAlpha = true
			} else {
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
		} else {
			context.SetSourceRGB(self.Config.Background.Red, self.Config.Background.Green, self.Config.Background.Blue)
		}

		context.SetOperator(cairo.OPERATOR_SOURCE)
		context.Paint()
	}
}

func (self *Window) updateWindowShapePixmapFromWebview() error {
	self.webview.GetSnapshotSurfaceWithOptions(webkit2.RegionFullDocument, webkit2.SnapshotOptionTransparentBackground, func(surface *cairo.Surface, err error) {
		if err == nil {
			if err := self.gtkWindow.SetShape(surface); err != nil {
				log.Errorf("Failed to reshape window: %v", err)
			}
		} else {
			log.Errorf("Failed to reshape window: %v", err)
		}
	})

	return nil
}
