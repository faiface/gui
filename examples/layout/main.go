package main

import (
	"image"
	"image/draw"
	"log"
	"time"

	"github.com/faiface/gui"
	"github.com/faiface/gui/fixedgrid"
	"github.com/faiface/gui/win"
	"github.com/faiface/mainthread"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/gofont/goregular"
)

func Blinker(env gui.Env, closed bool) {
	defer func() {
		if recover() != nil {
			log.Print("recovered blinker")
		}
	}()

	var r image.Rectangle
	var visible bool = true
	// redraw takes a bool and produces a draw command
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
	go func() {
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

	if closed {
		time.Sleep(time.Second * 1)
		close(env.Draw())
	}
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
	w, err := win.New(win.Title("gui test"),
		win.Resizable(),
	)
	if err != nil {
		panic(err)
	}
	mux, env := gui.NewMux(w)
	gr := fixedgrid.New(mux.MakeEnv(),
		fixedgrid.Rows(5),
		fixedgrid.Columns(2),
		fixedgrid.Gap(10),
	)
	log.Print(gr)
	go Blinker(gr.GetEnv("0;0"), false)
	go Blinker(gr.GetEnv("0;1"), true)
	go Blinker(gr.GetEnv("1;1"), false)
	go Blinker(gr.GetEnv("0;2"), false)
	go Blinker(gr.GetEnv("0;3"), false)
	go Blinker(gr.GetEnv("0;4"), false)
	sgr := fixedgrid.New(gr.GetEnv("1;0"),
		fixedgrid.Columns(3),
		fixedgrid.Gap(4),
		fixedgrid.Background(colornames.Darkgrey),
	)
	go Button(sgr.GetEnv("0;0"), theme, "Hey", func() {
		log.Print("hey")
	})
	go Button(sgr.GetEnv("1;0"), theme, "Ho", func() {
		log.Print("ho")
	})
	go Button(sgr.GetEnv("2;0"), theme, "Hu", func() {
		log.Print("hu")
	})
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
