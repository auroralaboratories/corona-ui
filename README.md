# Corona UI
A transparent, borderless Webkit frame (and not much else)

## Overview

The Corona UI is a very tiny web browser designed to load a single-page web application.  These applications will, in turn, talk to the [Corona API](https://github.com/auroralaboratories/corona-api) for performing desktop and system management tasks.  Collectively, these tasks form the necessary interactions and behaviors of using a modern Linux graphical environment.

## Command Line Usage

```
corona [options] APP
```

### `[options]`

| Argument Name         | Description                                                                                                                |
| --------------------- | -------------------------------------------------------------------------------------------------------------------------- |
| `--hide`              | Hide the window on startup, leaving it up to the application being launched to show it when it is ready.                   |
| `--show-in-panel`     | Show the window's icon in the system panel                                                                                 |
| `-L / --layer`        | Which layer of the window stacking order the window should be ordered in (desktop, below, **normal**, above)               |
| `-w / --width`        | The initial width of the window, in pixels (_e.g.: 250_) or percent of screen width (_e.g.: 75%_)                          |
| `-h / --height`       | The initial height of the window, in pixels (_e.g.: 32_) or percent of screen height (_e.g.: 5%_)                          |
| `-X`                  | The X-coordinate at which the window should be placed initially                                                            |
| `-Y`                  | The Y-coordinate at which the window should be placed initially                                                            |
| `-D / --dock`         | A shortcut for pinning the window to a particular edge of the screen (top, left, bottom, right)                            |
| `-A / --align`        | A shortcut for aligning the window within the axis the window is docked to (start, middle, end)                            |
| `-R / --reserve`      | Have this window reserve its dimensions so that other windows won't maximize over it.                                      |
| `-E / --decorator`    | Let window display window manager decorations.                                                                             |


### `APP`

The `APP` argument is mandatory and tells `corona` which page it should load.  If an absolute path is specified (that is, a path that starts with a _/_ ), then `corona` will attempt to load the file at that path.  If the absolute path is a directory, it will attempt to load the file `index.html` in that directory.  Otherwise, the value of `APP` is treated as an _application name_, and a series of directories will be searched to locate the application, with the first existing path being loaded.  The directories that are searched can be overridden by setting the `CORONA_PATH` environment variable.  The default search path is:

* _~/.corona/apps_
* _/usr/share/corona/apps_

The value of `APP` will be appended to each of these paths, then an _index.html_ file be loaded.  So, for example, given the command `corona paneltest`, this is the series of paths that would be searched:

```
~/.corona/apps/paneltest/index.html
/usr/share/corona/paneltest/index.html
```

With the first extant file being loaded.  If no file could be found, `corona` will exit immediately with a non-zero exit status.

