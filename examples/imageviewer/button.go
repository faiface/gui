package main

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/faiface/gui"
)

func Button(env gui.Env, theme *Theme, text string, action func()) {
	textImg := MakeTextImage(text, theme.Face, theme.Text)

	redraw := func(r image.Rectangle, over, pressed bool) func(draw.Image) image.Rectangle {
		return func(drw draw.Image) image.Rectangle {
			var clr color.Color
			if pressed {
				clr = theme.ButtonDown
			} else if over {
				clr = theme.ButtonOver
			} else {
				clr = theme.ButtonUp
			}
			draw.Draw(drw, r, &image.Uniform{clr}, image.ZP, draw.Src)
			DrawCentered(drw, r, textImg, draw.Over)
			return r
		}
	}

	var (
		r       image.Rectangle
		over    bool
		pressed bool
	)

	for e := range env.Events() {
		var x, y, x0, y0, x1, y1 int

		switch {
		case e.Matches("resize/%d/%d/%d/%d", &x0, &y0, &x1, &y1):
			r = image.Rect(x0, y0, x1, y1)
			env.Draw() <- redraw(r, over, pressed)

		case e.Matches("mo/down/%d/%d/left", &x, &y):
			newPressed := image.Pt(x, y).In(r)
			if newPressed != pressed {
				pressed = newPressed
				env.Draw() <- redraw(r, over, pressed)
			}

		case e.Matches("mo/up/%d/%d/left", &x, &y):
			if pressed {
				if image.Pt(x, y).In(r) {
					action()
				}
				pressed = false
				env.Draw() <- redraw(r, over, pressed)
			}
		}
	}

	close(env.Draw())
}
