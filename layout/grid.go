package layout

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/faiface/gui"
)

// Grid represents a simple grid layout.
// Do not edit properties directly, use the constructor instead.
type Grid struct {
	Contents   [][]*gui.Env
	Background color.Color
	Gap        int
	SplitX     func(int, int) []int
	SplitY     func(int, int) []int
}

func NewGrid(env gui.Env, contents [][]*gui.Env, options ...func(*Grid)) {
	ret := &Grid{
		Background: image.Black,
		Gap:        0,
		Contents:   contents,
		SplitX:     evenSplit,
		SplitY:     evenSplit,
	}
	for _, f := range options {
		f(ret)
	}

	mux := NewMux(env, ret)
	for _, row := range contents {
		for _, item := range row {
			*item, _ = mux.makeEnv(false)
		}
	}
}

func GridBackground(c color.Color) func(*Grid) {
	return func(grid *Grid) {
		grid.Background = c
	}
}

func GridGap(g int) func(*Grid) {
	return func(grid *Grid) {
		grid.Gap = g
	}
}

func GridSplitX(split func(int, int) []int) func(*Grid) {
	return func(grid *Grid) {
		grid.SplitX = split
	}
}

func GridSplitY(split func(int, int) []int) func(*Grid) {
	return func(grid *Grid) {
		grid.SplitY = split
	}
}

func (g *Grid) Redraw(drw draw.Image, bounds image.Rectangle) {
	draw.Draw(drw, bounds, image.NewUniform(g.Background), image.ZP, draw.Src)
}

func (g *Grid) Lay(bounds image.Rectangle) []image.Rectangle {
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
