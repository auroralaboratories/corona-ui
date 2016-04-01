// +build linux
// +build !no_x11

package gdk

// #cgo pkg-config: gdk-x11-3.0
// #include <gdk/gdk.h>
// #include <gdk/gdkx.h>
import "C"

// GetWindowID is a wrapper around gdk_x11_window_get_xid().
// It only works on GDK versions compiled with X11 support.  It will return an error if X11 support is unavailable
func (v *Window) GetWindowID() (uint32, error) {
	xid := C.gdk_x11_window_get_xid(v.native())
	return uint32(xid), nil
}
