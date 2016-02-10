// +build !linux no_x11

package gdk

import "fmt"

import "fmt"

func (v *Window) MoveToCurrentDesktop() {
}

// GetDesktop is a wrapper around gdk_x11_window_get_desktop().
// It only works on GDK versions compiled with X11 support - its return value can't be used if WorkspaceControlSupported returns false
func (v *Window) GetDesktop() uint32 {
	return 0
}

// MoveToDesktop is a wrapper around gdk_x11_window_move_to_desktop().
// It only works on GDK versions compiled with X11 support - its return value can't be used if WorkspaceControlSupported returns false
func (v *Window) MoveToDesktop(d uint32) {
}

// GetWindowID is a wrapper around gdk_x11_window_get_xid().
// It only works on GDK versions compiled with X11 support.  It will return an error if X11 support is unavailable
func (v *Window) GetWindowID() (uint32, error) {
    return 0, fmt.Errorf("Cannot retrieve window ID: X11 is unavailable")
}
