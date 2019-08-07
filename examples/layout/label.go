package main

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/faiface/gui"
)

func Label(env gui.Env, theme *Theme, text string, colr color.Color) {
	textImg := MakeTextImage(text, theme.Face, theme.Text)

	redraw := func(r image.Rectangle) func(draw.Image) image.Rectangle {
		return func(drw draw.Image) image.Rectangle {
			draw.Draw(drw, r, &image.Uniform{colr}, image.ZP, draw.Src)
			DrawLeftCentered(drw, r.Add(image.Pt(5, 0)), textImg, draw.Over)
			return r
		}
	}

	var (
		r image.Rectangle
	)

	for e := range env.Events() {
		switch e := e.(type) {
		case gui.Resize:
			r = e.Rectangle
			env.Draw() <- redraw(r)
		}
	}

	close(env.Draw())
}
