package layout

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/faiface/gui"
)

type grid struct {
	Contents   [][]*gui.Env
	Background color.Color
	Gap        int
	SplitX     SplitFunc
	SplitY     SplitFunc
}

// NewGrid creates a familiar flexbox-like grid layout.
// Each row can be a different length.
func NewGrid(env gui.Env, contents [][]*gui.Env, options ...func(*grid)) gui.Env {
	ret := &grid{
		Background: image.Black,
		Gap:        0,
		Contents:   contents,
		SplitX:     EvenSplit,
		SplitY:     EvenSplit,
	}
	for _, f := range options {
		f(ret)
	}

	mux, env := NewMux(env, ret)
	for _, row := range contents {
		for _, item := range row {
			*item = mux.MakeEnv()
		}
	}

	return env
}

// GridBackground changes the background of the grid to a uniform color.
func GridBackground(c color.Color) func(*grid) {
	return func(grid *grid) {
		grid.Background = c
	}
}

// GridGap changes the grid gap.
// The gap is identical everywhere (top, left, bottom, right).
func GridGap(g int) func(*grid) {
	return func(grid *grid) {
		grid.Gap = g
	}
}

// GridSplitX changes the way the space is divided among the columns in each row.
func GridSplitX(split SplitFunc) func(*grid) {
	return func(grid *grid) {
		grid.SplitX = split
	}
}

// GridSplitY changes the way the space is divided among the rows.
func GridSplitY(split SplitFunc) func(*grid) {
	return func(grid *grid) {
		grid.SplitY = split
	}
}

func (g *grid) Redraw(drw draw.Image, bounds image.Rectangle) {
	draw.Draw(drw, bounds, image.NewUniform(g.Background), image.ZP, draw.Src)
}

func (g *grid) Lay(bounds image.Rectangle) []image.Rectangle {
	gap := g.Gap
	ret := make([]image.Rectangle, 0)
	rows := len(g.Contents)

	rowsH := g.SplitY(rows, bounds.Dy()-(g.Gap*(rows+1)))

	X := gap + bounds.Min.X
	Y := gap + bounds.Min.Y
	for y, row := range g.Contents {
		cols := len(row)
		h := rowsH[y]
		colsW := g.SplitX(cols, bounds.Dx()-(g.Gap*(cols+1)))
		X = gap + bounds.Min.X
		for x := range row {
			w := colsW[x]
			ret = append(ret, image.Rect(X, Y, X+w, Y+h))
			X += gap + w
		}
		Y += gap + h
	}

	return ret
}
