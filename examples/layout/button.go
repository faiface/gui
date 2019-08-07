package main

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/faiface/gui"
	"github.com/faiface/gui/win"
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
		switch e := e.(type) {
		case gui.Resize:
			r = e.Rectangle
			env.Draw() <- redraw(r, over, pressed)

		case win.MoMove:
			nover := e.Point.In(r)
			if nover != over {
				over = nover
				env.Draw() <- redraw(r, over, pressed)
			}

		case win.MoDown:
			newPressed := e.Point.In(r)
			if newPressed != pressed {
				pressed = newPressed
				env.Draw() <- redraw(r, over, pressed)
			}

		case win.MoUp:
			if pressed {
				if e.Point.In(r) {
					action()
				}
				pressed = false
				env.Draw() <- redraw(r, over, pressed)
			}
		}
	}

	close(env.Draw())
}
