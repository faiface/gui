package main

import (
	"image"
	"image/draw"
	"log"
	"time"

	"github.com/faiface/gui"
	"github.com/faiface/gui/win"
	"golang.org/x/image/colornames"
)

func Blinker(env gui.Env) {
	defer func() {
		if recover() != nil {
			log.Print("recovered blinker")
		}
	}()

	var r image.Rectangle
	var visible bool = true

	redraw := func() func(draw.Image) image.Rectangle {
		return func(drw draw.Image) image.Rectangle {
			if visible {
				draw.Draw(drw, r, image.White, image.ZP, draw.Src)
			} else {
				draw.Draw(drw, r, &image.Uniform{colornames.Firebrick}, image.ZP, draw.Src)
			}
			return r
		}
	}

	// first we draw a white rectangle
	env.Draw() <- redraw()
	func() {
		for event := range env.Events() {
			switch event := event.(type) {
			case win.MoDown:
				if event.Point.In(r) {
					go func() {
						for i := 0; i < 3; i++ {
							visible = false
							env.Draw() <- redraw()
							time.Sleep(time.Second / 3)
							visible = true
							env.Draw() <- redraw()
							time.Sleep(time.Second / 3)
						}
					}()
				}
			case gui.Resize:
				log.Print(event)
				r = event.Rectangle
				env.Draw() <- redraw()
			}
		}
	}()
}
