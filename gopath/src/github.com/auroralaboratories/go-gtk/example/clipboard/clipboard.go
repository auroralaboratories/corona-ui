package main

import (
	"fmt"
	"github.com/auroralaboratories/go-gtk/gdk"
	"github.com/auroralaboratories/go-gtk/gtk"
)

func main() {
	gtk.Init(nil)
	clipboard := gtk.NewClipboardGetForDisplay(
		gdk.DisplayGetDefault(),
		gdk.SELECTION_CLIPBOARD)
	fmt.Println(clipboard.WaitForText())
	clipboard.SetText("helloworld")
	gtk.MainIterationDo(true)
	clipboard.Store()
	gtk.MainIterationDo(true)
}
