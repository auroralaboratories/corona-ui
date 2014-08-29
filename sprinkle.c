#include <gtk/gtk.h>
#include <gdk/gdkscreen.h>
#include <cairo.h>
#include <webkit/webkit.h>

gboolean supports_alpha = FALSE;

static void screen_changed(GtkWidget *widget, GdkScreen *old_screen, gpointer user_data);
static gboolean expose(GtkWidget *widget, GdkEventExpose *event, gpointer user_data);
static void clicked(GtkWindow *win, GdkEventButton *event, gpointer user_data);

static void destroy_cb(GtkWidget* widget, gpointer data) {
  gtk_main_quit();
}

int main(int argc, char* argv[]) {
  gtk_init(&argc, &argv);

//create main window and Webkit widget
  GtkWidget           *window = gtk_window_new(GTK_WINDOW_TOPLEVEL);
  WebKitWebView     *web_view = WEBKIT_WEB_VIEW(webkit_web_view_new());
  WebKitWebSettings *settings = webkit_web_settings_new();

  // maximize
  //gtk_window_maximize(GTK_WINDOW(window));

  // set intitial size
  gtk_window_set_default_size(GTK_WINDOW(window), 512, 512);


  // pin to desktop
  //gtk_window_set_type_hint(GTK_WINDOW(window), GDK_WINDOW_TYPE_HINT_DESKTOP);


  // callback: quit GTK mainloop
  g_signal_connect(window, "destroy",        G_CALLBACK(destroy_cb), NULL);

  // callback(s): handle cairo double buffering (this is what enables the top-level to be transparent)
  g_signal_connect(window, "expose-event",   G_CALLBACK(expose), NULL);
  g_signal_connect(window, "screen-changed", G_CALLBACK(screen_changed), NULL);

  // callback(s): handle cairo double buffering (this is what enables the WEBKIT WIDGET to be transparent)
  g_signal_connect(web_view, "expose-event",   G_CALLBACK(expose), NULL);
  g_signal_connect(web_view, "screen-changed", G_CALLBACK(screen_changed), NULL);

  // disable titlebar and border
  gtk_window_set_decorated(GTK_WINDOW(window), FALSE);

  // do this for reasons
  gtk_widget_set_app_paintable(window, TRUE);

  // enable <audio> tag
  g_object_set (G_OBJECT(settings), "auto-load-images",                  TRUE, NULL);


  // enable extensions
  g_object_set (G_OBJECT(settings), "enable-plugins",                    TRUE, NULL);

  // enable javascript
  g_object_set (G_OBJECT(settings), "enable-scripts",                    TRUE, NULL);

  // enable HW acceleration
  g_object_set (G_OBJECT(settings), "enable-accelerated-compositing",    TRUE, NULL);

  // enable WebGL
  g_object_set (G_OBJECT(settings), "enable-webgl",                      TRUE, NULL);

  // enable <audio> tag
  g_object_set (G_OBJECT(settings), "enable-webaudio",                   TRUE, NULL);

  // do this too
  g_object_set (G_OBJECT(settings), "enable-file-access-from-file-uris", TRUE, NULL);


  // set the settings from above
  webkit_web_view_set_settings (WEBKIT_WEB_VIEW(web_view), settings);

  // enable fully transparent backgrounds
  webkit_web_view_set_transparent(WEBKIT_WEB_VIEW(web_view), TRUE);


  gtk_container_add(GTK_CONTAINER(window), GTK_WIDGET(web_view));

  // first CLI argument is the page to load, otherwise go to blank
  if(argc > 1){
    webkit_web_view_load_uri(web_view, argv[1]);
  }else{
    webkit_web_view_load_uri(web_view, "about:blank");
  }

  // focus the window
  gtk_widget_grab_focus(GTK_WIDGET(web_view));

  // do initial double buffer
  screen_changed(window, NULL, NULL);

  // show the window
  gtk_widget_show_all(window);

  // enter mainloop (blocks until destroy)
  gtk_main();

  return 0;
}


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
        supports_alpha = TRUE;
    }

    gtk_widget_set_colormap(widget, colormap);
}

static gboolean expose(GtkWidget *widget, GdkEventExpose *event, gpointer userdata)
{
   cairo_t *cr = gdk_cairo_create(widget->window);

    if (supports_alpha){
        cairo_set_source_rgba (cr, 1.0, 1.0, 1.0, 0.0); /* transparent */
    }else{
        cairo_set_source_rgb (cr, 1.0, 0.5, 1.0); /* opaque white */
    }

    /* draw the background */
    cairo_set_operator(cr, CAIRO_OPERATOR_SOURCE);
    cairo_paint(cr);
    cairo_destroy(cr);

    return FALSE;
}
