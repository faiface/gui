package main

import (
	"github.com/faiface/gui"
	"github.com/faiface/gui/layout"
	"golang.org/x/image/colornames"
)

func Card(env gui.Env, theme *Theme, title, content string) {
	box := layout.Grid{
		Rows: []int{1, 1},
		// Flip:       true,
		// Gap:        4,
		Background: colornames.Pink,
	}
	fields := makeEnvPtr(2)
	layout.NewMux(env,
		fields,
		box,
	)
	go Label(*fields[0], theme, title, colornames.Lightgray)
	go Label(*fields[1], theme, content, colornames.Slategray)
	// go Blinker(*fields[1])
}
