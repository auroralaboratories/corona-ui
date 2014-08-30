#include <gtk/gtk.h>
#include <gdk/gdkscreen.h>
#include <cairo.h>
#include <webkit/webkit.h>

#define SP_WM_TYPE_DESKTOP  "desktop"
#define SP_WM_TYPE_DOCK     "dock"

#define SP_WM_LAYER_BELOW   "below"
#define SP_WM_LAYER_NORMAL  "normal"
#define SP_WM_LAYER_ABOVE   "above"

#define SP_WM_DOCK_TOP      "top"
#define SP_WM_DOCK_LEFT     "left"
#define SP_WM_DOCK_BOTTOM   "bottom"
#define SP_WM_DOCK_RIGHT    "right"

#define SP_WM_ALIGN_START   "start"
#define SP_WM_ALIGN_MIDDLE  "middle"
#define SP_WM_ALIGN_END     "end"

gboolean supports_alpha = FALSE;

static void screen_changed(GtkWidget *widget, GdkScreen *old_screen, gpointer user_data);
static gboolean expose(GtkWidget *widget, GdkEventExpose *event, gpointer user_data);
static void clicked(GtkWindow *win, GdkEventButton *event, gpointer user_data);
void sprinkle_apply_flags(GtkWindow *window);

static void destroy_cb(GtkWidget* widget, gpointer data) {
  gtk_main_quit();
}

static gboolean start_hidden  = FALSE;
static gboolean show_in_panel = FALSE;
static gchar*   wm_type       = 0;
static gchar*   wm_layer      = SP_WM_LAYER_NORMAL;
static guint    wm_width      = 0;
static guint    wm_height     = 0;
static gint     wm_xpos       = 0;
static gint     wm_ypos       = 0;
static gchar*   wm_dock       = NULL;
static gchar*   wm_align      = NULL;
static gboolean wm_autostrut  = FALSE;

static GOptionEntry entries[] =
{
  { "hide",          0,   0, G_OPTION_ARG_NONE,   &start_hidden,  "Hide the window on startup, leaving it up to the application being launched to show it when it is ready", NULL },
  { "show-in-panel", 0,   0, G_OPTION_ARG_NONE,   &show_in_panel, "Show the window's icon in the system panel", NULL },
  { "type",          'T', 0, G_OPTION_ARG_STRING, &wm_type,       "What type of window should this be flagged as (desktop, dock)", NULL },
  { "layer",         'L', 0, G_OPTION_ARG_STRING, &wm_layer,      "Which layer of the window stacking order the window should be ordered in", SP_WM_LAYER_NORMAL },
  { "width",         'w', 0, G_OPTION_ARG_INT,    &wm_width,      "Initial width of the window, in pixels", NULL },
  { "height",        'h', 0, G_OPTION_ARG_INT,    &wm_height,     "Initial height of the window, in pixels", NULL },
  { "xpos",          'X', 0, G_OPTION_ARG_INT,    &wm_xpos,       "The X-coordinate at which the window should be placed initially", NULL },
  { "ypos",          'Y', 0, G_OPTION_ARG_INT,    &wm_ypos,       "The Y-coordinate at which the window should be placed initially", NULL },
  { "dock",          'D', 0, G_OPTION_ARG_STRING, &wm_dock,       "A shortcut for pinning the window to a particular edge of the screen (top, left, bottom, right)", NULL},
  { "align",         'A', 0, G_OPTION_ARG_STRING, &wm_align,      "A shortcut for aligning the window within the axis the window is docked to (start, middle, end)", NULL},
  { "reserve",       'R', 0, G_OPTION_ARG_NONE,   &wm_autostrut,  "Have this window reserve its dimensions so that other windows won't maximize over it", NULL},
  { NULL }
};

int main(int argc, char* argv[]) {
  GError *error = NULL;
  GOptionContext *context;

  context = g_option_context_new("- lightweight webkit browser");
  g_option_context_add_main_entries(context, entries, NULL);
  g_option_context_add_group(context, gtk_get_option_group(TRUE));
  if (!g_option_context_parse(context, &argc, &argv, &error)){
    g_print("option parsing failed: %s\n", error->message);
    return 1;
  }

//  gtk_init(&argc, &argv);

  g_print("Initializing...\n");
  printf("[%s] = %d\n", "hide",              start_hidden);
  printf("[%s] = %d\n", "show-in-panel",     show_in_panel);
  printf("[%s] = %s\n", "layer",             wm_layer);
  printf("[dimensions] = %dx%d @ (%d,%d)\n", wm_width, wm_height, wm_xpos, wm_ypos);
  printf("[%s] = %s\n", "dock",              wm_dock);
  printf("[%s] = %s\n", "align",             wm_align);
  printf("[%s] = %d\n", "reserve",           wm_autostrut);

//create main window and Webkit widget
  GtkWidget           *window = gtk_window_new(GTK_WINDOW_TOPLEVEL);

  gtk_window_set_default_size(GTK_WINDOW(window), 512, 512);


//apply the WM flags to the window
  sprinkle_apply_flags(GTK_WINDOW(window));

  WebKitWebView     *web_view = WEBKIT_WEB_VIEW(webkit_web_view_new());
  WebKitWebSettings *settings = webkit_web_settings_new();

  // maximize
  //gtk_window_maximize(GTK_WINDOW(window));

  // set intitial size


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
    g_print("Must specify an application name\n");
    return 127;
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


void sprinkle_apply_flags(GtkWindow *window) {
  GdkWindow *gdk_window = gtk_widget_get_window(GTK_WIDGET(window));

  if(start_hidden){
    gtk_widget_hide(GTK_WIDGET(window));
  }

  if(show_in_panel){
    gtk_window_set_skip_taskbar_hint(window, FALSE);
    gtk_window_set_skip_pager_hint(window, FALSE);
  }else{
    gtk_window_set_skip_taskbar_hint(window, TRUE);
    gtk_window_set_skip_pager_hint(window, TRUE);
  }

//TYPE
  if(wm_type == SP_WM_TYPE_DESKTOP){
    gtk_window_set_type_hint(window, GDK_WINDOW_TYPE_HINT_DESKTOP);
  }else if(wm_type == SP_WM_TYPE_DOCK){
    gtk_window_set_type_hint(window, GDK_WINDOW_TYPE_HINT_DOCK);
  }

//LAYER
  if(wm_layer == SP_WM_LAYER_BELOW){
    gdk_window_set_keep_below(gdk_window, TRUE);
  }else if(wm_layer == SP_WM_LAYER_ABOVE){
    gdk_window_set_keep_above(gdk_window, TRUE);
  }

  if(wm_width && wm_height) {
    gtk_window_resize(window, wm_width, wm_height);
  }

  if(wm_xpos && wm_ypos) {
    gtk_window_move(window, wm_xpos, wm_ypos);
  }else{
    GdkScreen *gdk_screen = gtk_window_get_screen(window);
    gint x, y, window_w, window_h = 01;

    gtk_window_get_size(window, &window_w, &window_h);

    g_print("Window current size: %dx%d\n", window_w, window_h);
    g_print("Screen is %dx%d\n", gdk_screen_get_width(gdk_screen), gdk_screen_get_height(gdk_screen));

//  set Y-coordinates
    if(wm_dock == SP_WM_DOCK_BOTTOM){
      y = gdk_screen_get_height(gdk_screen) - window_h;
    }else if(wm_dock == SP_WM_DOCK_RIGHT){
      y = gdk_screen_get_width(gdk_screen) - window_w;
    }

//  set X-coordinates
    if(!strcmp(wm_dock, SP_WM_DOCK_TOP) || !strcmp(wm_dock, SP_WM_DOCK_BOTTOM)){
      if(!strcmp(wm_align, SP_WM_ALIGN_MIDDLE)){
        x = (gdk_screen_get_width(gdk_screen) / 2.0) - (window_w / 2.0);
      }else if(!strcmp(wm_align,SP_WM_ALIGN_END)){
        x = gdk_screen_get_width(gdk_screen) - window_w;
      }
    }else if(!strcmp(wm_dock, SP_WM_DOCK_LEFT) || !strcmp(wm_dock, SP_WM_DOCK_RIGHT)){
      if(!strcmp(wm_align, SP_WM_ALIGN_MIDDLE)){
        x = (gdk_screen_get_height(gdk_screen) / 2.0) - (window_h / 2.0);
      }else if(!strcmp(wm_align, SP_WM_ALIGN_END)){
        x = gdk_screen_get_height(gdk_screen) - window_h;
      }
    }

    g_print("Moving window to %d, %d\n", x, y);

    gtk_window_move(window, x, y);
  }

}