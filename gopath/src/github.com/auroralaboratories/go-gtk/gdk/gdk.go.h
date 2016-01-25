#ifndef GO_GDK_H
#define GO_GDK_H

#include <gdk/gdk.h>
#include <cairo/cairo.h>
#include <unistd.h>
#include <stdlib.h>
#include <stdint.h>
#include <stdarg.h>
#include <string.h>

static gchar* toGstr(char* s) { return (gchar*)s; }

static void freeCstr(char* s) { free(s); }

static GdkWindow* toGdkWindow(void* w) { return GDK_WINDOW(w); }
static GdkDragContext* toGdkDragContext(void* l) { return (GdkDragContext*)l; }
static GdkScreen* toGdkScreen(void* s) { return GDK_SCREEN(s); }
static GdkColormap* toGdkColormap(void* c) { return GDK_COLORMAP(c); }
static cairo_content_t toCairoContentT(int type){
  switch(type){
  case CAIRO_CONTENT_COLOR:
    return CAIRO_CONTENT_COLOR;
  case CAIRO_CONTENT_ALPHA:
    return CAIRO_CONTENT_ALPHA;
  default:
    return CAIRO_CONTENT_COLOR_ALPHA;
  }
}

static void* _gdk_display_get_default() {
  return (void*) gdk_display_get_default();
}


#endif
