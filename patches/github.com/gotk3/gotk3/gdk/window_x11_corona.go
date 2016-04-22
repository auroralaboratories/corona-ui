// +build linux
// +build !no_x11

package gdk

// #cgo pkg-config: gdk-x11-3.0
// #include <gdk/gdk.h>
// #include <gdk/gdkx.h>
import "C"

type WindowTypeHint int

const (
	TypeHintNormal        WindowTypeHint = C.GDK_WINDOW_TYPE_HINT_NORMAL        // Normal toplevel window.
	TypeHintDialog                       = C.GDK_WINDOW_TYPE_HINT_DIALOG        // Dialog window.
	TypeHintMenu                         = C.GDK_WINDOW_TYPE_HINT_MENU          // Window used to implement a menu; GTK+ uses this hint only for torn-off menus, see GtkTearoffMenuItem.
	TypeHintToolbar                      = C.GDK_WINDOW_TYPE_HINT_TOOLBAR       // Window used to implement toolbars.
	TypeHintSplashscreen                 = C.GDK_WINDOW_TYPE_HINT_SPLASHSCREEN  // Window used to display a splash screen during application startup.
	TypeHintUtility                      = C.GDK_WINDOW_TYPE_HINT_UTILITY       // Utility windows which are not detached toolbars or dialogs.
	TypeHintDock                         = C.GDK_WINDOW_TYPE_HINT_DOCK          // Used for creating dock or panel windows.
	TypeHintDesktop                      = C.GDK_WINDOW_TYPE_HINT_DESKTOP       // Used for creating the desktop background window.
	TypeHintDropdown_menu                = C.GDK_WINDOW_TYPE_HINT_DROPDOWN_MENU // A menu that belongs to a menubar.
	TypeHintPopup_menu                   = C.GDK_WINDOW_TYPE_HINT_POPUP_MENU    // A menu that does not belong to a menubar, e.g. a context menu.
	TypeHintTooltip                      = C.GDK_WINDOW_TYPE_HINT_TOOLTIP       // A tooltip.
	TypeHintNotification                 = C.GDK_WINDOW_TYPE_HINT_NOTIFICATION  // A notification - typically a “bubble” that belongs to a status icon.
	TypeHintCombo                        = C.GDK_WINDOW_TYPE_HINT_COMBO         // A popup from a combo box.
	TypeHintDnd                          = C.GDK_WINDOW_TYPE_HINT_DND           // A window that is used to implement a DND cursor.
)

// GetWindowID is a wrapper around gdk_x11_window_get_xid().
// It only works on GDK versions compiled with X11 support.  It will return an error if X11 support is unavailable
func (v *Window) GetWindowID() (uint32, error) {
	xid := C.gdk_x11_window_get_xid(v.native())
	return uint32(xid), nil
}
