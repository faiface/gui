package main

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/faiface/gui"
	"github.com/faiface/gui/win"
	"github.com/faiface/mainthread"
	"github.com/fogleman/gg"
)

func ColorPicker(env gui.Env, pick chan<- color.Color, r image.Rectangle, clr color.Color) {
	env.Draw() <- func(drw draw.Image) image.Rectangle {
		draw.Draw(drw, r, &image.Uniform{clr}, r.Min, draw.Src)
		return r
	}

	for event := range env.Events() {
		var x, y int
		switch {
		case event.Matches("mo/down/%d/%d", &x, &y):
			if image.Pt(x, y).In(r) {
				pick <- clr
			}
		}
	}

	close(env.Draw())
}

func Canvas(env gui.Env, pick <-chan color.Color, r image.Rectangle) {
	canvas := image.NewRGBA(r)
	draw.Draw(canvas, r, image.White, r.Min, draw.Src)
	dc := gg.NewContextForRGBA(canvas)

	env.Draw() <- func(drw draw.Image) image.Rectangle {
		draw.Draw(drw, r, canvas, r.Min, draw.Src)
		return r
	}

	var (
		clr     = color.Color(color.Black)
		pressed = false
		px, py  = 0, 0
	)

	for {
		select {
		case clr = <-pick:

		case event, ok := <-env.Events():
			if !ok {
				close(env.Draw())
				return
			}

			var x, y int
			switch {
			case event.Matches("mo/down/%d/%d", &x, &y):
				if image.Pt(x, y).In(r) {
					pressed = true
					px, py = x, y
				}

			case event.Matches("mo/up/%d/%d", &x, &y):
				pressed = false

			case event.Matches("mo/move/%d/%d", &x, &y):
				if pressed {
					x0, y0, x1, y1 := px, py, x, y
					px, py = x, y

					env.Draw() <- func(drw draw.Image) image.Rectangle {
						dc.SetColor(clr)
						dc.SetLineCapRound()
						dc.SetLineWidth(5)
						dc.DrawLine(float64(x0), float64(y0), float64(x1), float64(y1))
						dc.Stroke()
						rect := image.Rect(x0, y0, x1, y1)
						rect.Min.X -= 5
						rect.Min.Y -= 5
						rect.Max.X += 5
						rect.Max.Y += 5
						draw.Draw(drw, rect, canvas, rect.Min, draw.Src)
						return rect
					}
				}
			}
		}
	}
}

func run() {
	w, err := win.New(win.Title("Paint"), win.Size(800, 600))
	if err != nil {
		panic(err)
	}

	mux, env := gui.NewMux(w)

	pick := make(chan color.Color)

	for i, clr := range []color.Color{
		color.RGBA{255, 0, 0, 255},
		color.RGBA{255, 255, 0, 255},
		color.RGBA{0, 255, 0, 255},
		color.RGBA{0, 255, 255, 255},
		color.RGBA{0, 0, 255, 255},
		color.RGBA{255, 0, 255, 255},
		color.RGBA{255, 255, 255, 255},
		color.RGBA{0, 0, 0, 255},
	} {
		go ColorPicker(mux.MakeEnv(), pick, image.Rect(750, i*75, 800, (i+1)*75), clr)
	}

	go Canvas(mux.MakeEnv(), pick, image.Rect(0, 0, 750, 600))

	for event := range env.Events() {
		switch {
		case event.Matches("wi/close"):
			close(env.Draw())
			return
		}
	}
}

func main() {
	mainthread.Run(run)
}
