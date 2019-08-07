package main

import (
	"image"
	"image/color"
	"image/draw"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/faiface/gui"
	"github.com/faiface/gui/win"
)

func Blinker(env gui.Env) {
	defer func() {
		if recover() != nil {
			log.Print("recovered blinker")
		}
	}()
	buf := make([]byte, 3)
	rand.Read(buf)
	defaultColor := image.NewUniform(color.RGBA{buf[0], buf[1], buf[2], 255})
	rand.Read(buf)
	blinkColor := image.NewUniform(color.RGBA{buf[0], buf[1], buf[2], 255})
	redraw := func(r image.Rectangle, visible bool) func(draw.Image) image.Rectangle {
		return func(drw draw.Image) image.Rectangle {
			if r == image.ZR {
				return r
			}
			if visible {
				draw.Draw(drw, r, defaultColor, image.ZP, draw.Src)
			} else {
				draw.Draw(drw, r, blinkColor, image.ZP, draw.Src)
			}
			return r
		}
	}

	var mu sync.Mutex
	var (
		r       image.Rectangle
		visible bool = true
	)

	// first we draw a white rectangle
	// env.Draw() <- redraw(b)
	func() {
		for event := range env.Events() {
			switch event := event.(type) {
			case win.MoDown:
				if event.Point.In(r) {
					go func() {
						for i := 0; i < 6; i++ {
							mu.Lock()
							visible = !visible
							env.Draw() <- redraw(r, visible)
							mu.Unlock()

							time.Sleep(time.Second / 3)
						}
					}()
				}
			case gui.Resize:
				r = event.Rectangle
				env.Draw() <- redraw(r, visible)
			}
		}
	}()
}
