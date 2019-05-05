package main

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/faiface/gui"
)

func Button(env gui.Env, theme *Theme, text string, action func()) {
	textImg := DrawText(text, theme.Face, theme.Text)

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
		var x, y, minX, minY, maxX, maxY int

		switch {
		case e.Matches("resize/%d/%d/%d/%d", &minX, &minY, &maxX, &maxY):
			r = image.Rect(minX, minY, maxX, maxY)
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
