package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math/rand"
	"time"

	"github.com/faiface/gui"
	"github.com/faiface/gui/win"
	"github.com/faiface/mainthread"
)

func EqualColors(c1, c2 color.Color) bool {
	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()
	return r1 == r2 && g1 == g2 && b1 == b2 && a1 == a2
}

func HexToColor(hex string) color.Color {
	var r, g, b uint8
	fmt.Sscanf(hex, "#%2X%2X%2X", &r, &g, &b)
	return color.RGBA{r, g, b, 255}
}

var Colors = []color.Color{
	HexToColor("#E53935"),
	HexToColor("#F06292"),
	HexToColor("#9C27B0"),
	HexToColor("#673AB7"),
	HexToColor("#3F51B5"),
	HexToColor("#2196F3"),
	HexToColor("#29B6F6"),
	HexToColor("#00BCD4"),
	HexToColor("#009688"),
	HexToColor("#4CAF50"),
	HexToColor("#8BC34A"),
	HexToColor("#CDDC39"),
	HexToColor("#FFEB3B"),
	HexToColor("#FFC107"),
	HexToColor("#FF9800"),
	HexToColor("#8D6E63"),
	HexToColor("#9E9E9E"),
	HexToColor("#607D8B"),
}

type PairMsg struct {
	Color color.Color
	Resp  chan<- bool
}

func Tile(env gui.Env, pair chan PairMsg, r image.Rectangle, clr color.Color) {
	redraw := func(covered float64) func(draw.Image) image.Rectangle {
		return func(drw draw.Image) image.Rectangle {
			coveredY := int(float64(r.Dy()) * covered)
			bottomR := r
			bottomR.Min.Y = bottomR.Max.Y - coveredY
			topR := r
			topR.Max.Y = bottomR.Min.Y
			draw.Draw(drw, bottomR, &image.Uniform{HexToColor("#37474F")}, image.ZP, draw.Src)
			draw.Draw(drw, topR, &image.Uniform{clr}, image.ZP, draw.Src)
			return r
		}
	}

	env.Draw() <- redraw(0.0)

	for event := range env.Events() {
		var x, y int
		switch {
		case event.Matches("mo/down/%d/%d", &x, &y):
			if image.Pt(x, y).In(r) {
				for c := 32; c >= 0; c-- {
					env.Draw() <- redraw(float64(c) / 32)
					time.Sleep(time.Second / 32 / 4)
				}

				var correct bool

				resp := make(chan bool)
				select {
				case pair <- PairMsg{clr, resp}:
					correct = <-resp

				case msg := <-pair:
					time.Sleep(time.Second / 2)
					correct = EqualColors(clr, msg.Color)
					msg.Resp <- correct
				}

				if correct {
					close(env.Draw())
					return
				}

				for c := 0; c <= 32; c++ {
					env.Draw() <- redraw(float64(c) / 32)
					time.Sleep(time.Second / 32 / 4)
				}
			}
		}
	}
}

func run() {
	rand.Seed(time.Now().UnixNano())

	w, err := win.New(win.Title("Pexeso"), win.Size(600, 600))
	if err != nil {
		panic(err)
	}

	mux, env := gui.NewMux(w)

	var colors []color.Color
	for _, clr := range Colors {
		colors = append(colors, clr, clr)
	}

	rand.Shuffle(len(colors), func(i, j int) {
		colors[i], colors[j] = colors[j], colors[i]
	})

	env.Draw() <- func(drw draw.Image) image.Rectangle {
		r := image.Rect(0, 0, 600, 600)
		draw.Draw(drw, r, &image.Uniform{HexToColor("#CFD8DC")}, image.ZP, draw.Src)
		return r
	}

	pair := make(chan PairMsg)

	i := 0
	for x := 0; x < 600; x += 100 {
		for y := 0; y < 600; y += 100 {
			go Tile(mux.MakeEnv(), pair, image.Rect(
				x+10, y+10,
				x+90, y+90,
			), colors[i])
			i++
		}
	}

	for event := range env.Events() {
		switch {
		case event.Matches("wi/close"):
			close(env.Draw())
		}
	}
}

func main() {
	mainthread.Run(run)
}
