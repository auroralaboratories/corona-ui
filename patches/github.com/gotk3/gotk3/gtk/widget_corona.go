package gtk

// #include <gtk/gtk.h>
// #include "gtk.go.h"
import "C"
import (
	"unsafe"

	"github.com/gotk3/gotk3/gdk"
)

// GetScreen is a wrapper around gtk_widget_get_screen().
func (v *Widget) GetScreen() (*gdk.Screen, error) {
	c := C.gtk_widget_get_screen(v.native())
	if c == nil {
		return nil, nilPtrErr
	}

	w := &gdk.Screen{wrapObject(unsafe.Pointer(c))}
	return w, nil
}
