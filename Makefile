all:
	gcc sprinkle.c `pkg-config --cflags --libs gtk+-2.0 webkit-1.0 cairo x11` -o bin/sprinkle

panel-tint2:
	./sprinkle file://${PWD}/apps/panel-tint2/panel.html
