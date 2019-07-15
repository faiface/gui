package main

import (
	"image"
	"image/draw"
	"log"
	"time"

	"github.com/faiface/gui"
	"github.com/faiface/gui/layout"
	"github.com/faiface/gui/win"
	"github.com/faiface/mainthread"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/gofont/goregular"
)

func run() {
	face, err := TTFToFace(goregular.TTF, 18)
	if err != nil {
		panic(err)
	}
	theme := &Theme{
		Face:       face,
		Background: colornames.White,
		Empty:      colornames.Darkgrey,
		Text:       colornames.Black,
		Highlight:  colornames.Blueviolet,
		ButtonUp:   colornames.Lightgrey,
		ButtonDown: colornames.Grey,
	}
	w, err := win.New(win.Title("gui test")) // win.Resizable(),
	if err != nil {
		panic(err)
	}

	mux, env := gui.NewMux(w)

	go func() {
		// Hack for non-reparenting window managers (I think)
		e := mux.MakeEnv()
		for {
			time.Sleep(time.Second / 5)
			e.Draw() <- func(drw draw.Image) image.Rectangle {
				r := image.Rect(0, 0, 10, 10)
				draw.Draw(drw, r, image.Transparent, image.ZP, draw.Over)
				return r
			}
		}
	}()

	var (
		top                             gui.Env
		left, right                     gui.Env
		bottomLeft, bottom, bottomRight gui.Env
	)
	layout.NewMux(
		mux.MakeEnv(),
		[]*gui.Env{
			&top,
			&left, &right,
			&bottomLeft, &bottom, &bottomRight},
		layout.Grid{
			Rows:       []int{1, 2, 3},
			Gap:        10,
			Background: colornames.Sandybrown,
			SplitY: func(els int, width int) []int {
				ret := make([]int, els)
				total := 0
				for i := 0; i < els; i++ {
					if i == els-1 {
						ret[i] = width - total
					} else {
						v := (width - total) / 2
						ret[i] = v
						total += v
					}
				}
				return ret
			},
		},
	)
	go Blinker(right)
	go Blinker(left)
	go Blinker(bottomRight)

	var (
		b1, b2, b3, b4, b5, b6 gui.Env
	)
	layout.NewMux(top,
		[]*gui.Env{&b1, &b2, &b3},
		layout.Box{
			Length:     3,
			Gap:        10,
			Background: colornames.Lightblue,
		},
	)
	go Blinker(b1)
	go Blinker(b2)
	box := layout.Box{
		Length:     3,
		Vertical:   true,
		Gap:        4,
		Background: colornames.Pink,
		Split: func(els int, width int) []int {
			ret := make([]int, els)
			total := 0
			for i := 0; i < els-1; i++ {
				v := (width - total) / 2
				ret[i] = v
				total += v
			}
			ret[els-1] = width - total
			return ret
		},
	}

	layout.NewMux(b3,
		[]*gui.Env{
			&b4, &b5, &b6,
		},
		box,
	)

	go Blinker(b4)
	go Blinker(b5)
	go Blinker(b6)

	var (
		btn1, btn2, btn3 gui.Env
	)
	layout.NewMux(
		bottom,
		[]*gui.Env{&btn1, &btn2, &btn3},
		layout.Grid{
			Rows:       []int{2, 1},
			Background: colornames.Darkgrey,
			Gap:        4,
		},
	)
	btn := func(env gui.Env, name string) {
		Button(env, theme, name, func() {
			log.Print(name)
		})
	}
	go btn(btn1, "Hey")
	go btn(btn2, "Ho")
	go btn(btn3, "Hu")

	// we use the master env now, w is used by the mux
	for event := range env.Events() {
		switch event.(type) {
		case win.WiClose:
			close(env.Draw())
		}
	}
}

func main() {
	mainthread.Run(run)
}
