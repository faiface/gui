package main

import (
	"fmt"
	"image"
	"image/draw"
	"log"
	"os"
	"time"

	"github.com/faiface/gui"
	"github.com/faiface/gui/layout"
	"github.com/faiface/gui/win"
	"github.com/faiface/mainthread"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/gofont/goregular"
)

func makeEnvPtr(n int) []*gui.Env {
	elsp := make([]*gui.Env, n)
	for i := 0; i < len(elsp); i++ {
		elsp[i] = new(gui.Env)
	}
	return elsp
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
		ButtonOver: colornames.Grey,
		ButtonDown: colornames.Dimgrey,
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
			time.Sleep(time.Second / 10)
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
			Rows:   []int{1, 2, 3},
			Gap:    10,
			Margin: -6,
			Border: 1,
			// Flip:        true,
			BorderColor: image.White,
			Background:  colornames.Sandybrown,
			SplitRows: func(els int, width int) []int {
				ret := make([]int, els)
				total := 0
				for i := 0; i < els-1; i++ {
					ret[i] = (width - total) / 2
					total += ret[i]
				}
				ret[els-1] = width - total
				return ret
			},
		},
	)
	go Blinker(right)
	go Blinker(left)
	go Blinker(bottomRight)

	subGrid := makeEnvPtr(3)
	layout.NewMux(top,
		subGrid,
		layout.Grid{
			Rows:       []int{len(subGrid)},
			Gap:        10,
			Background: colornames.Lightblue,
		},
	)

	elsp := makeEnvPtr(100)
	scrl := &layout.Scroller{
		Background:  colornames.Red,
		Length:      len(elsp),
		Gap:         2,
		ChildHeight: 80,
	}
	layout.NewMux(*subGrid[0],
		elsp,
		scrl,
	)
	for i, el := range elsp {
		// go Blinker(*el)
		go Card(*el, theme, "hello", fmt.Sprintf("I'm card #%d", i))
	}

	go Blinker(*subGrid[1])
	box := layout.Grid{
		Rows:       []int{3},
		Flip:       true,
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
	blinkers := makeEnvPtr(3)
	layout.NewMux(*subGrid[2],
		blinkers,
		box,
	)

	go Blinker(*blinkers[0])
	go Blinker(*blinkers[1])
	go Blinker(*blinkers[2])

	btns := makeEnvPtr(3)
	layout.NewMux(
		bottom,
		btns,
		layout.Grid{
			Rows:       []int{2, 1},
			Background: colornames.Darkgrey,
			Gap:        4,
			Flip:       true,
		},
	)
	btn := func(env gui.Env, name string) {
		Button(env, theme, name, func() {
			log.Print(name)
		})
	}
	go btn(*btns[0], "Hey")
	go btn(*btns[1], "Ho")
	go btn(*btns[2], "Hu")

	// we use the master env now, w is used by the mux
	for event := range env.Events() {
		switch event.(type) {
		case win.WiClose:
			close(env.Draw())
			os.Exit(0)
		}
	}
}

func main() {
	mainthread.Run(run)
}
