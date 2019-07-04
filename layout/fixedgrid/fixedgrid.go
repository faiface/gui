package fixedgrid

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"github.com/faiface/gui"
	"github.com/faiface/gui/layout"
)

type FixedGrid struct {
	Columns    int
	Rows       int
	Background color.Color
	Gap        int

	*layout.Layout
}

func New(env gui.Env, options ...func(*FixedGrid)) *FixedGrid {
	ret := &FixedGrid{
		// Bounds:     image.ZR,
		Background: image.Black,
		Columns:    1,
		Rows:       1,
		Gap:        0,
	}

	for _, f := range options {
		f(ret)
	}

	ret.Layout = layout.New(env, ret.layout, ret.redraw)
	return ret
}

func (g *FixedGrid) layout(bounds image.Rectangle) map[string]image.Rectangle {
	gap := g.Gap
	cols := g.Columns
	rows := g.Rows

	w := (bounds.Dx() - (cols+1)*gap) / cols
	h := (bounds.Dy() - (rows+1)*gap) / rows

	ret := make(map[string]image.Rectangle)
	X := gap + bounds.Min.X
	Y := gap + bounds.Min.Y
	for x := 0; x < cols; x++ {
		for y := 0; y < rows; y++ {
			ret[fmt.Sprintf("%d;%d", x, y)] = image.Rect(X, Y, X+w, Y+h)
			Y += gap + h
		}
		Y = gap + bounds.Min.Y
		X += gap + w
	}

	return ret
}

func Background(c color.Color) func(*FixedGrid) {
	return func(grid *FixedGrid) {
		grid.Background = c
	}
}

func Gap(g int) func(*FixedGrid) {
	return func(grid *FixedGrid) {
		grid.Gap = g
	}
}

func Columns(cols int) func(*FixedGrid) {
	return func(grid *FixedGrid) {
		grid.Columns = cols
	}
}

func Rows(rows int) func(*FixedGrid) {
	return func(grid *FixedGrid) {
		grid.Rows = rows
	}
}

func (g *FixedGrid) redraw(drw draw.Image, bounds image.Rectangle) {
	draw.Draw(drw, bounds, image.NewUniform(g.Background), image.ZP, draw.Src)
}
