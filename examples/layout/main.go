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
	layout.NewGrid(
		mux.MakeEnv(),
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
	)
	go Blinker(right)
	go Blinker(left)
	go Blinker(bottomRight)

	var (
		b1, b2, b3, b4, b5, b6 gui.Env
	)
	layout.NewBox(
		top,
		[]*gui.Env{
			&b1, &b2, &b3,
		},
		layout.BoxGap(10),
		layout.BoxBackground(colornames.Lightblue),
	)
	go Blinker(b1)
	go Blinker(b2)
	layout.NewBox(
		b3,
		[]*gui.Env{
			&b4, &b5, &b6,
		},
		layout.BoxVertical,
		layout.BoxBackground(colornames.Pink),
		layout.BoxGap(4),
		layout.BoxSplit(func(els int, width int) []int {
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
	)
	go Blinker(b4)
	go Blinker(b5)
	go Blinker(b6)

	var (
		btn1, btn2, btn3 gui.Env
	)
	layout.NewGrid(
		bottom,
		[][]*gui.Env{
			{&btn1, &btn2, &btn3},
		},
		layout.GridGap(4),
		layout.GridBackground(colornames.Darkgrey),
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
