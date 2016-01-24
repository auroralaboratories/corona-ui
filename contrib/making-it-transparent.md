# Implementing transparent GTK window backgrounds
Via: http://stackoverflow.com/a/3909283

```c
#include <gtk/gtk.h>
#include <gdk/gdkscreen.h>
#include <cairo.h>

static void screen_changed(GtkWidget *widget, GdkScreen *old_screen, gpointer user_data);
static gboolean expose(GtkWidget *widget, GdkEventExpose *event, gpointer user_data);
static void clicked(GtkWindow *win, GdkEventButton *event, gpointer user_data);

int main(int argc, char **argv)
{
    gtk_init(&argc, &argv);

    GtkWidget *window = gtk_window_new(GTK_WINDOW_TOPLEVEL);
    gtk_window_set_position(GTK_WINDOW(window), GTK_WIN_POS_CENTER);
    gtk_window_set_default_size(GTK_WINDOW(window), 400, 400);
    gtk_window_set_title(GTK_WINDOW(window), "Alpha Demo");
    g_signal_connect(G_OBJECT(window), "delete-event", gtk_main_quit, NULL);

    gtk_widget_set_app_paintable(window, TRUE);

    g_signal_connect(G_OBJECT(window), "expose-event", G_CALLBACK(expose), NULL);
    g_signal_connect(G_OBJECT(window), "screen-changed", G_CALLBACK(screen_changed), NULL);

    gtk_window_set_decorated(GTK_WINDOW(window), FALSE);
    gtk_widget_add_events(window, GDK_BUTTON_PRESS_MASK);
    g_signal_connect(G_OBJECT(window), "button-press-event", G_CALLBACK(clicked), NULL);

    GtkWidget* fixed_container = gtk_fixed_new();
    gtk_container_add(GTK_CONTAINER(window), fixed_container);
    GtkWidget* button = gtk_button_new_with_label("button1");
    gtk_widget_set_size_request(button, 100, 100);
    gtk_container_add(GTK_CONTAINER(fixed_container), button);

    screen_changed(window, NULL, NULL);

    gtk_widget_show_all(window);
    gtk_main();

    return 0;
}


gboolean supports_alpha = FALSE;
static void screen_changed(GtkWidget *widget, GdkScreen *old_screen, gpointer userdata)
{
    /* To check if the display supports alpha channels, get the colormap */
    GdkScreen *screen = gtk_widget_get_screen(widget);
    GdkColormap *colormap = gdk_screen_get_rgba_colormap(screen);

    if (!colormap)
    {
        printf("Your screen does not support alpha channels!\n");
        colormap = gdk_screen_get_rgb_colormap(screen);
        supports_alpha = FALSE;
    }
    else
    {
        printf("Your screen supports alpha channels!\n");
        supports_alpha = TRUE;
    }

    gtk_widget_set_colormap(widget, colormap);
}

static gboolean expose(GtkWidget *widget, GdkEventExpose *event, gpointer userdata)
{
   cairo_t *cr = gdk_cairo_create(widget->window);

    if (supports_alpha)
        cairo_set_source_rgba (cr, 1.0, 1.0, 1.0, 0.0); /* transparent */
    else
        cairo_set_source_rgb (cr, 1.0, 1.0, 1.0); /* opaque white */

    /* draw the background */
    cairo_set_operator (cr, CAIRO_OPERATOR_SOURCE);
    cairo_paint (cr);

    cairo_destroy(cr);

    return FALSE;
}

static void clicked(GtkWindow *win, GdkEventButton *event, gpointer user_data)
{
    /* toggle window manager frames */
    gtk_window_set_decorated(win, !gtk_window_get_decorated(win));
}
```

## Overview

1. Tell GTK to allow our code to paint the background. This prevents GTK from
   applying the theming on its own.

2. Connect to the 'screen-changed' and 'expose-event' signals.  The callbacks
   that the application implements will be responsible for the background drawing.

3. expose-event implementation: use GDK + Cairo to paint the background, which does
   the necessary magic to draw the underlying screen into the GTK window background.

   This means that the GTK window background is in fact a Cairo surface.

4. screen-changed handles things like turning on/off a compositor and colorspace
   changes that would affect our ability to do transparency.


## Necessary Functions

### gtk2
* **gtk_widget_set_app_paintable**

  Exists: gtk.SetAppPaintable()
* **gtk_window_set_decorated**

  Exists: gtk.SetDecorated()
* **gtk_widget_get_screen**

  NOT IMPLEMENTED: gtk.go@9600
* **gtk_widget_set_colormap**

  NOT IMPLEMENTED, gtk.go@9488

### gdk2
```
gdk_screen_get_rgba_colormap    MISSING, see: https://developer.gnome.org/gdk2/stable/GdkScreen.html#gdk-screen-get-rgba-colormap
gdk_screen_get_rgb_colormap     MISSING, see: https://developer.gnome.org/gdk2/stable/GdkScreen.html#gdk-screen-get-rgb-colormap
gdk_cairo_create                MISSING, see: https://developer.gnome.org/gdk2/stable/gdk2-Cairo-Interaction.html#gdk-cairo-create
```

### cairo
```
cairo_set_source_rgba           MISSING, see: http://www.cairographics.org/manual/cairo-cairo-t.html#cairo-set-source-rgba
cairo_set_source_rgb            MISSING, see: http://www.cairographics.org/manual/cairo-cairo-t.html#cairo-set-source-rgb
cairo_set_operator              MISSING, see: http://www.cairographics.org/manual/cairo-cairo-t.html#cairo-set-operator
cairo_paint                     MISSING, see: http://www.cairographics.org/manual/cairo-cairo-t.html#cairo-paint
cairo_destroy                   MISSING, see: http://www.cairographics.org/manual/cairo-cairo-t.html#cairo-destroy
```