// Same copyright and license as the rest of the files in this project
// This file contains accelerator related functions and structures
package gtk

// #include <cairo.h>
// #include <gtk/gtk.h>
// #include "gtk.go.h"
import "C"
import (
	"fmt"
	"unsafe"

	"github.com/gotk3/gotk3/cairo"
)

func (v *Window) SetShape(surface *cairo.Surface) error {
	if gdkScreen, err := v.GetScreen(); err == nil {
		if gdkDisplay, err := gdkScreen.GetDisplay(); err == nil {
			if !gdkDisplay.SupportsShapes() {
				return fmt.Errorf("Underlying display does not support arbitrary window shapes")
			}
		} else {
			return err
		}
	} else {
		return err
	}

	if surface == nil {
		C.gtk_widget_shape_combine_region(v.toWidget(), nil)
	} else {
		var region *C.cairo_region_t

		region = C.gdk_cairo_region_create_from_surface((*C.cairo_surface_t)(unsafe.Pointer(surface.Native())))

		if region != nil {
			C.gtk_widget_shape_combine_region(v.toWidget(), region)
		}
	}

	return nil
}
