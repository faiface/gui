package main

import (
	"log"

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
	var (
		top                             gui.Env
		left, right                     gui.Env
		bottomLeft, bottom, bottomRight gui.Env
	)
	layout.NewMux(
		mux.MakeEnv(),
		layout.NewGrid(
			[][]*gui.Env{
				{&top},
				{&left, &right},
				{&bottomLeft, &bottom, &bottomRight},
			},
			layout.GridGap(10),
			layout.GridBackground(colornames.Sandybrown),
			layout.GridSplitY(func(els int, width int) []int {
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
			}),
		),
	)
	go Blinker(right)
	go Blinker(left)
	go Blinker(bottomRight)

	var (
		b1, b2, b3, b4, b5, b6 gui.Env
	)
	layout.NewMux(top,
		layout.NewBox(
			[]*gui.Env{
				&b1, &b2, &b3,
			},
			layout.BoxGap(10),
			layout.BoxBackground(colornames.Lightblue),
		),
	)
	go Blinker(b1)
	go Blinker(b2)
	box := layout.NewBox(
		[]*gui.Env{
			&b4, &b5, &b6,
		},
		layout.BoxVertical,
		layout.BoxBackground(colornames.Pink),
		layout.BoxGap(4),
		layout.BoxSplit(func(els int, width int) []int {
			ret := make([]int, els)
			total := 0
			for i := 0; i < els-1; i++ {
				v := (width - total) / 2
				ret[i] = v
				total += v
			}
			ret[els-1] = width - total
			return ret
		}),
	)
	layout.NewMux(b3, box)
	log.Print(box)

	go Blinker(b4)
	go Blinker(b5)
	go Blinker(b6)

	var (
		btn1, btn2, btn3 gui.Env
	)
	layout.NewMux(
		bottom,
		layout.NewGrid(
			[][]*gui.Env{
				{&btn1, &btn2, &btn3},
			},
			layout.GridGap(4),
			layout.GridBackground(colornames.Darkgrey),
		),
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
