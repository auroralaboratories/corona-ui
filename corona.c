#include <cairo.h>
#include <gdk/gdkscreen.h>
#include <gtk/gtk.h>
#include <math.h>
#include <pixman.h>
#include <stdlib.h>
#include <string.h>
#include <sys/stat.h>
#include <sys/types.h>
#include <unistd.h>
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

static gchar* corona_find_application_by_name(gchar *name);
static gchar* corona_application_path(gchar *path, gchar *name);
static gchar* corona_path_suffix_index(gchar *uri);
static void on_show(GtkWidget *widget, gpointer user_data);
static gboolean on_popup_window(WebKitWebView             *web_view,
               WebKitWebFrame            *frame,
               WebKitNetworkRequest      *request,
               WebKitWebNavigationAction *navigation_action,
               WebKitWebPolicyDecision   *policy_decision,
               gpointer                   user_data);

static void screen_changed(GtkWidget *widget, GdkScreen *old_screen, gpointer user_data);
static gboolean expose(GtkWidget *widget, GdkEventExpose *event, gpointer user_data);
static void clicked(GtkWindow *win, GdkEventButton *event, gpointer user_data);
static cairo_surface_t* blur_image_surface(cairo_surface_t *surface, int radius, double sigma);
static pixman_fixed_t* create_gaussian_blur_kernel(int radius, double sigma, int *length);
static gboolean navigate(
  WebKitWebView             *web_view,
  WebKitWebFrame            *frame,
  WebKitNetworkRequest      *request,
  WebKitWebNavigationAction *navigation_action,
  WebKitWebPolicyDecision   *policy_decision,
  gpointer                   user_data);

void corona_apply_flags(GtkWindow *window);

static void destroy_cb(GtkWidget* widget, gpointer data) {
  gtk_main_quit();
}

static gboolean  start_hidden  = FALSE;
static gboolean  show_in_panel = FALSE;
static gchar*    wm_type       = 0;
static gchar*    wm_layer      = SP_WM_LAYER_NORMAL;
static gint      wm_width      = -1;
static gint      wm_height     = -1;
static gint      wm_xpos       = -1;
static gint      wm_ypos       = -1;
static gchar*    wm_dock       = NULL;
static gchar*    wm_align      = NULL;
static gboolean  wm_autostrut  = FALSE;
static gboolean  wm_root_win   = FALSE;
static gboolean  wm_decorator  = FALSE;
static gchar*    sp_system     = "/usr/share/corona/apps";
static gchar*    sp_user       = "~/.corona/apps";

static GOptionEntry entries[] =
{
  { "hide",          0,   0, G_OPTION_ARG_NONE,   &start_hidden,  "Hide the window on startup, leaving it up to the application being launched to show it when it is ready", NULL },
  { "show-in-panel", 0,   0, G_OPTION_ARG_NONE,   &show_in_panel, "Show the window's icon in the system panel", NULL },
  { "type",          'T', 0, G_OPTION_ARG_STRING, &wm_type,       "What type of window should this be flagged as (desktop, dock)", NULL },
  { "layer",         'L', 0, G_OPTION_ARG_STRING, &wm_layer,      "Which layer of the window stacking order the window should be ordered in", SP_WM_LAYER_NORMAL },
  { "width",         'w', 0, G_OPTION_ARG_INT   , &wm_width,      "Initial width of the window, in pixels", NULL },
  { "height",        'h', 0, G_OPTION_ARG_INT   , &wm_height,     "Initial height of the window, in pixels", NULL },
  { "xpos",          'X', 0, G_OPTION_ARG_INT   , &wm_xpos,       "The X-coordinate at which the window should be placed initially", NULL },
  { "ypos",          'Y', 0, G_OPTION_ARG_INT   , &wm_ypos,       "The Y-coordinate at which the window should be placed initially", NULL },
  { "dock",          'D', 0, G_OPTION_ARG_STRING, &wm_dock,       "A shortcut for pinning the window to a particular edge of the screen (top, left, bottom, right)", NULL},
  { "align",         'A', 0, G_OPTION_ARG_STRING, &wm_align,      "A shortcut for aligning the window within the axis the window is docked to (start, middle, end)", NULL},
  { "root",          'r', 0, G_OPTION_ARG_NONE,   &wm_root_win,   "Make this window's parent the root window (draw on desktop)", NULL},
  { "reserve",       'R', 0, G_OPTION_ARG_NONE,   &wm_autostrut,  "Have this window reserve its dimensions so that other windows won't maximize over it", NULL},
  { "decorator",     'E', 0, G_OPTION_ARG_NONE,   &wm_decorator,  "Let window display window manager decorations", NULL},
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
  printf("[%s] = %d\n", "decorator",         wm_decorator);

//create main window and Webkit widget
  GtkWidget           *window = gtk_window_new(GTK_WINDOW_TOPLEVEL);
  GtkWidget           *layout = gtk_scrolled_window_new(NULL, NULL);

  gtk_window_set_default_size(GTK_WINDOW(window), 512, 512);

  WebKitWebView           *web_view = WEBKIT_WEB_VIEW(webkit_web_view_new());
  WebKitWebSettings       *settings = webkit_web_settings_new();
  WebKitWebWindowFeatures *features = webkit_web_view_get_window_features(web_view);

  // maximize
  //gtk_window_maximize(GTK_WINDOW(window));

  // set intitial size


  // pin to desktop
  //gtk_window_set_type_hint(GTK_WINDOW(window), GDK_WINDOW_TYPE_HINT_DESKTOP);


  // callback: quit GTK mainloop
  g_signal_connect(window,   "destroy",        G_CALLBACK(destroy_cb), NULL);
  g_signal_connect(window,   "show",           G_CALLBACK(on_show), NULL);

  // callback(s): handle cairo double buffering (this is what enables the top-level to be transparent)
  g_signal_connect(window,   "expose-event",   G_CALLBACK(expose), NULL);
  g_signal_connect(window,   "screen-changed", G_CALLBACK(screen_changed), NULL);

  // callback(s): handle cairo double buffering
  g_signal_connect(layout,   "expose-event",   G_CALLBACK(expose), NULL);
  g_signal_connect(layout,   "screen-changed", G_CALLBACK(screen_changed), NULL);

  // callback(s): handle cairo double buffering (this is what enables the WEBKIT WIDGET to be transparent)
  g_signal_connect(web_view, "expose-event",   G_CALLBACK(expose), NULL);
  g_signal_connect(web_view, "screen-changed", G_CALLBACK(screen_changed), NULL);

  g_signal_connect(web_view, "navigation-policy-decision-requested", G_CALLBACK(navigate), NULL);
  g_signal_connect(web_view, "new-window-policy-decision-requested", G_CALLBACK(on_popup_window), NULL);


  g_object_set(G_OBJECT(web_view), "self-scrolling",                     TRUE, NULL);


  // disable titlebar and border
  gtk_window_set_decorated(GTK_WINDOW(window), wm_decorator);

  // do this for reasons
  gtk_widget_set_app_paintable(window, TRUE);
  gtk_widget_set_app_paintable(layout, TRUE);

  // enable images
  g_object_set (G_OBJECT(settings), "auto-load-images",                  TRUE, NULL);

  // disable auto-resize
  g_object_set (G_OBJECT(settings), "auto-resize-window",                FALSE, NULL);

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


  // set window features
  g_object_set(G_OBJECT(features), "locationbar-visible",                FALSE, NULL);
  g_object_set(G_OBJECT(features), "menubar-visible",                    FALSE, NULL);
  g_object_set(G_OBJECT(features), "scrollbar-visible",                  FALSE, NULL);
  g_object_set(G_OBJECT(features), "statusbar-visible",                  FALSE, NULL);
  g_object_set(G_OBJECT(features), "toolbar-visible",                    FALSE, NULL);



  // set the settings from above
  webkit_web_view_set_settings (WEBKIT_WEB_VIEW(web_view), settings);

  // enable fully transparent backgrounds
  webkit_web_view_set_transparent(WEBKIT_WEB_VIEW(web_view), TRUE);

  gtk_container_add(GTK_CONTAINER(layout), GTK_WIDGET(web_view));
  gtk_container_add(GTK_CONTAINER(window), GTK_WIDGET(layout));

  // first CLI argument is the page to load, otherwise go to blank
  if(argc > 1){
    gchar *uri = g_strdup(argv[1]);

    if(!g_str_has_prefix(uri, "http") && !g_str_has_prefix(uri, "file")){
      if(g_str_has_prefix(uri, "/")){
        uri = corona_path_suffix_index(uri);
      }else{
        uri = corona_find_application_by_name(uri);

        if(uri == NULL){
          g_print("Could not find application '%s'\n", argv[1]);
          return 64;
        }

        uri = corona_path_suffix_index(uri);
      }

      uri = g_strdup_printf("file://%s", uri);
    }

    webkit_web_view_load_uri(web_view, uri);
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

static gchar* corona_find_application_by_name(gchar *name){
  gchar *rv;

  if((rv = corona_application_path("./apps", name)) != NULL){
    return rv;
  }

  if((rv = corona_application_path(sp_user, name)) != NULL){
    return rv;
  }

  if((rv = corona_application_path(sp_system, name)) != NULL){
    return rv;
  }

  return NULL;
}

static gchar* corona_application_path(gchar *path, gchar *name){
  struct stat buffer;
  gchar       *exp_path = NULL;
  gchar       *filename = NULL;

  exp_path = realpath(path, NULL);

  if(exp_path == NULL){
    filename = g_strdup_printf("%s/%s", path, name);
  }else{
    filename = g_strdup_printf("%s/%s", exp_path, name);
  }

  g_print("Search %s\n", filename);

  if (stat(filename, &buffer) == 0){
    g_print("File found! %s\n", filename);
    return filename;
  }

  return NULL;
}


static gchar* corona_path_suffix_index(gchar *uri){
  if(!g_str_has_suffix(uri, ".html")){
    uri = g_strconcat(uri, "/index.html", NULL);
  }

  return uri;
}

static void on_show(GtkWidget *widget, gpointer userdata){
  //apply the WM flags to the window
    corona_apply_flags(GTK_WINDOW(widget));
}

static gboolean on_popup_window(WebKitWebView             *web_view,
               WebKitWebFrame            *frame,
               WebKitNetworkRequest      *request,
               WebKitWebNavigationAction *navigation_action,
               WebKitWebPolicyDecision   *policy_decision,
               gpointer                   user_data)
{
  webkit_web_policy_decision_ignore(policy_decision);
  return TRUE;
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
   cairo_surface_t *surface = cairo_get_target(cr);

    if (supports_alpha){
        cairo_set_source_rgba (cr, 1.0, 1.0, 1.0, 0.0); /* transparent */
    }else{
        cairo_set_source_rgb (cr, 1.0, 0.5, 1.0); /* opaque white */
    }

    //blur_image_surface(surface, 2, 3.0);

    /* draw the background */
    cairo_set_operator(cr, CAIRO_OPERATOR_SOURCE);
    cairo_paint(cr);
    cairo_destroy(cr);

    return FALSE;
}

static gboolean navigate(
  WebKitWebView             *web_view,
  WebKitWebFrame            *frame,
  WebKitNetworkRequest      *request,
  WebKitWebNavigationAction *navigation_action,
  WebKitWebPolicyDecision   *policy_decision,
  gpointer                   user_data)
{
  g_print("URI: %s\n", webkit_network_request_get_uri(request));
  return FALSE;
}

void corona_apply_flags(GtkWindow *window) {
  GdkWindow *gdk_window = gtk_widget_get_window(GTK_WIDGET(window));
  GdkScreen *gdk_screen = gtk_window_get_screen(window);
  GdkWindow *gdk_root   = gdk_screen_get_root_window(gdk_screen);

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
  if(!g_strcmp0(wm_type,SP_WM_TYPE_DESKTOP)){
    gtk_window_set_type_hint(window, GDK_WINDOW_TYPE_HINT_DESKTOP);
  }else if(!g_strcmp0(wm_type,SP_WM_TYPE_DOCK)){
    gtk_window_set_type_hint(window, GDK_WINDOW_TYPE_HINT_DOCK);
  }

//LAYER
  if(!g_strcmp0(wm_layer,SP_WM_LAYER_BELOW)){
    gdk_window_set_keep_below(gdk_window, TRUE);
  }else if(!g_strcmp0(wm_layer, SP_WM_LAYER_ABOVE)){
    gdk_window_set_keep_above(gdk_window, TRUE);
  }

  gint window_w = 0;
  gint window_h = 0;
  gint window_x = 0;
  gint window_y = 0;

  if(wm_width && wm_height) {
    window_w = wm_width;
    window_h = wm_height;
    gdk_window_resize(gdk_window, wm_width, wm_height);
  }else{
    gdk_window_get_geometry(gdk_window, NULL, NULL, &window_w, &window_h, NULL);
  }

  if(wm_xpos >= 0 && wm_ypos >= 0) {
    gdk_window_move(gdk_window, wm_xpos, wm_ypos);
  }else{
    g_print("Window current size: %dx%d\n", window_w, window_h);
    g_print("Screen is %dx%d\n", gdk_screen_get_width(gdk_screen), gdk_screen_get_height(gdk_screen));

//  set Y-coordinates
    if(!g_strcmp0(wm_dock,SP_WM_DOCK_BOTTOM)){
      window_y = gdk_screen_get_height(gdk_screen) - window_h;

    }else if(!g_strcmp0(wm_dock,SP_WM_DOCK_RIGHT)){
      window_y = gdk_screen_get_width(gdk_screen) - window_w;
    }

//  set X-coordinates
    if(!g_strcmp0(wm_dock, SP_WM_DOCK_TOP) || !g_strcmp0(wm_dock, SP_WM_DOCK_BOTTOM)){
      if(!g_strcmp0(wm_align, SP_WM_ALIGN_MIDDLE)){
        window_x = (gdk_screen_get_width(gdk_screen) / 2.0) - (window_w / 2.0);
      }else if(!g_strcmp0(wm_align,SP_WM_ALIGN_END)){
        window_x = gdk_screen_get_width(gdk_screen) - window_w;
      }
    }else if(!g_strcmp0(wm_dock, SP_WM_DOCK_LEFT) || !g_strcmp0(wm_dock, SP_WM_DOCK_RIGHT)){
      if(!g_strcmp0(wm_align, SP_WM_ALIGN_MIDDLE)){
        window_x = (gdk_screen_get_height(gdk_screen) / 2.0) - (window_h / 2.0);
      }else if(!g_strcmp0(wm_align, SP_WM_ALIGN_END)){
        window_x = gdk_screen_get_height(gdk_screen) - window_h;
      }
    }

    g_print("Moving window to %d, %d\n", window_x, window_y);

    gdk_window_move(gdk_window, window_x, window_y);
  }

  gdk_window_get_geometry(gdk_window, &window_x, &window_y, NULL, NULL, NULL);

  if(wm_root_win){
    gdk_window_reparent(gdk_window, gdk_get_default_root_window(), window_x, window_y);
  }

//RESERVE
  if(wm_autostrut){
    GdkAtom atom;
    GdkAtom cardinal;
    unsigned long strut[12] = {0,0,0,0,0,0,0,0,0,0,0,0};

    if(!g_strcmp0(wm_dock, SP_WM_DOCK_TOP)){
      g_print("Reserving %d pixels at the top of the screen\n", window_h);

      strut[2]  = window_h;            // strut top
      strut[8]  = window_x;            // top_start_x
      strut[9]  = window_x + window_w; // top_end_x
    }else if(!g_strcmp0(wm_dock, SP_WM_DOCK_BOTTOM)){
      g_print("Reserving %d pixels at the bottom of the screen\n", window_h);

      strut[3]  = window_h;            // strut bottom
      strut[10] = window_x;            // bottom_start_x
      strut[11] = window_x + window_w; // bottom_end_x
    }else if(!g_strcmp0(wm_dock, SP_WM_DOCK_LEFT)){
      g_print("Reserving %d pixels on the left side of the screen\n", window_w);

      strut[0]  = window_w;            // strut left
      strut[4]  = window_y;            // left_start_y
      strut[5]  = window_y + window_h; // left_end_y
    }else if(!g_strcmp0(wm_dock, SP_WM_DOCK_RIGHT)){
      g_print("Reserving %d pixels on the right side of the screen\n", window_w);

      strut[1]  = window_w;            // strut right
      strut[6]  = window_y;            // right_start_y
      strut[7]  = window_y + window_h; // right_end_y
    }

    cardinal = gdk_atom_intern("CARDINAL", FALSE);
    atom = gdk_atom_intern("_NET_WM_STRUT", FALSE);

    gdk_property_change(gdk_window, atom, cardinal, 32, GDK_PROP_MODE_REPLACE,
        (guchar*)strut, 4);

    atom = gdk_atom_intern("_NET_WM_STRUT_PARTIAL", FALSE);
    gdk_property_change(gdk_window, atom, cardinal, 32, GDK_PROP_MODE_REPLACE,
        (guchar*)strut, 12);

    gdk_window_move(gdk_window, window_x, window_y);
  }
}


static pixman_fixed_t *
create_gaussian_blur_kernel (int     radius,
                             double  sigma,
                             int    *length)
{
  const double scale2 = 2.0 * sigma * sigma;
  const double scale1 = 1.0 / (M_PI * scale2);

  const int size = 2 * radius + 1;
  const int n_params = size * size;

  pixman_fixed_t *params;
  double *tmp, sum;
  int x, y, i;

  tmp = g_newa (double, n_params);

  /* caluclate gaussian kernel in floating point format */
  for (i = 0, sum = 0, x = -radius; x <= radius; ++x) {
          for (y = -radius; y <= radius; ++y, ++i) {
                  const double u = x * x;
                  const double v = y * y;

                  tmp[i] = scale1 * exp (-(u+v)/scale2);

                  sum += tmp[i];
          }
  }

  /* normalize gaussian kernel and convert to fixed point format */
  params = g_new (pixman_fixed_t, n_params + 2);

  params[0] = pixman_int_to_fixed (size);
  params[1] = pixman_int_to_fixed (size);

  for (i = 0; i < n_params; ++i)
          params[2 + i] = pixman_double_to_fixed (tmp[i] / sum);

  if (length)
          *length = n_params + 2;

  return params;
}

static cairo_surface_t* blur_image_surface(
  cairo_surface_t *surface,
  int              radius,
  double           sigma)
{
  static cairo_user_data_key_t data_key;
  pixman_fixed_t *params = NULL;
  int n_params;

  pixman_image_t *src, *dst;
  int w, h, s;
  gpointer p;

  g_return_val_if_fail
    (cairo_surface_get_type (surface) != CAIRO_SURFACE_TYPE_IMAGE,
     NULL);

  w = cairo_image_surface_get_width (surface);
  h = cairo_image_surface_get_height (surface);
  s = cairo_image_surface_get_stride (surface);

  /* create pixman image for cairo image surface */
  p = cairo_image_surface_get_data (surface);
  src = pixman_image_create_bits (PIXMAN_a8, w, h, p, s);

  /* attach gaussian kernel to pixman image */
  params = create_gaussian_blur_kernel(radius, sigma, &n_params);
  pixman_image_set_filter (src, PIXMAN_FILTER_CONVOLUTION, params, n_params);
  g_free (params);

  /* render blured image to new pixman image */
  p = g_malloc0 (s * h);
  dst = pixman_image_create_bits (PIXMAN_a8, w, h, p, s);
  pixman_image_composite (PIXMAN_OP_SRC, src, NULL, dst, 0, 0, 0, 0, 0, 0, w, h);
  pixman_image_unref (src);

  /* create new cairo image for blured pixman image */
  surface = cairo_image_surface_create_for_data (p, CAIRO_FORMAT_A8, w, h, s);
  cairo_surface_set_user_data (surface, &data_key, p, g_free);
  pixman_image_unref (dst);

  return surface;
}
