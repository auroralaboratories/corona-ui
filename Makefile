WARNINGS = -Wall -Wextra -Wno-unused-parameter -Wmissing-prototypes -Werror -Wfatal-errors
CFLAGS   = $(WARNINGS) -g -O2

all:
	[ ! -d bin ] && mkdir bin || true
	gcc sprinkle.c `pkg-config --cflags --libs gtk+-2.0 webkit-1.0 cairo x11 pixman-1` -lm -ldl -o bin/sprinkle

panel-tint2:
	./bin/sprinkle --width 1920 --height 24 --dock top --align middle file://${PWD}/apps/panel-tint2/index.html
