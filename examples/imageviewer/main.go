package main

import (
	"os/user"

	"github.com/faiface/gui"
	"github.com/faiface/gui/win"
	"github.com/faiface/mainthread"
	"github.com/golang/image/font/gofont/goregular"
	"golang.org/x/image/colornames"
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

	w, err := win.New(win.Title("Image Viewer"), win.Size(900, 600), win.Resizable())
	if err != nil {
		panic(err)
	}

	mux, env := gui.NewMux(w)

	cd := make(chan string)
	view := make(chan string)

	go Browser(FixedBottom(FixedLeft(mux.MakeEnv(), 300), 30), theme, ".", cd, view)
	go Viewer(FixedRight(mux.MakeEnv(), 300), theme, view)

	go Button(EvenHorizontal(FixedTop(FixedLeft(mux.MakeEnv(), 300), 30), 0, 1, 3), theme, "Dir Up", func() {
		cd <- ".."
	})
	go Button(EvenHorizontal(FixedTop(FixedLeft(mux.MakeEnv(), 300), 30), 1, 2, 3), theme, "Refresh", func() {
		cd <- "."
	})
	go Button(EvenHorizontal(FixedTop(FixedLeft(mux.MakeEnv(), 300), 30), 2, 3, 3), theme, "Home", func() {
		user, err := user.Current()
		if err != nil {
			return
		}
		cd <- user.HomeDir
	})

	for e := range env.Events() {
		switch {
		case e.Matches("wi/close"):
			close(env.Draw())
		}
	}
}

func main() {
	mainthread.Run(run)
}
