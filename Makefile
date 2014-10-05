WARNINGS = -Wall -Wextra -Wno-unused-parameter -Wmissing-prototypes -Werror -Wfatal-errors
CFLAGS   = $(WARNINGS) -g -O2

all:
	[ ! -d bin ] && mkdir bin || true
	gcc corona.c `pkg-config --cflags libconfig --libs gtk+-2.0 webkit-1.0 cairo x11 pixman-1 libconfig` -lm -ldl -o bin/corona

panel-tint2:
	./bin/corona --width 1920 --height 24 --dock top --align middle file://${PWD}/apps/panel-tint2/index.html
