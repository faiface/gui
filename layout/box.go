package layout

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/faiface/gui"
)

func evenSplit(elements int, width int) []int {
	ret := make([]int, 0, elements)
	for elements > 0 {
		v := width / elements
		width -= v
		elements -= 1
		ret = append(ret, v)
	}
	return ret
}

type Box struct {
	// Defaults to []*gui.Env{}
	Contents []*gui.Env
	// Defaults to image.Black
	Background color.Color
	// Defaults to an even split
	Split func(int, int) []int
	// Defaults to 0
	Gap int

	vertical bool
}

func NewBox(env gui.Env, contents []*gui.Env, options ...func(*Box)) gui.Env {
	ret := &Box{
		Background: image.Black,
		Contents:   contents,
		Split:      evenSplit,
	}
	for _, f := range options {
		f(ret)
	}

	mux, env := NewMux(env, ret)
	for _, item := range contents {
		*item = mux.MakeEnv()
	}
	return env
}

func BoxVertical(b *Box) {
	b.vertical = true
}

func BoxBackground(c color.Color) func(*Box) {
	return func(grid *Box) {
		grid.Background = c
	}
}

func BoxSplit(split func(int, int) []int) func(*Box) {
	return func(grid *Box) {
		grid.Split = split
	}
}

func BoxGap(gap int) func(*Box) {
	return func(grid *Box) {
		grid.Gap = gap
	}
}

func (g *Box) Redraw(drw draw.Image, bounds image.Rectangle) {
	draw.Draw(drw, bounds, image.NewUniform(g.Background), image.ZP, draw.Src)
}

func (g *Box) Lay(bounds image.Rectangle) []image.Rectangle {
	items := len(g.Contents)
	ret := make([]image.Rectangle, 0, items)
	if g.vertical {
		spl := g.Split(items, bounds.Dy()-(g.Gap*(items+1)))
		Y := bounds.Min.Y + g.Gap
		for _, item := range spl {
			ret = append(ret, image.Rect(bounds.Min.X+g.Gap, Y, bounds.Max.X-g.Gap, Y+item))
			Y += item + g.Gap
		}
	} else {
		spl := g.Split(items, bounds.Dx()-(g.Gap*(items+1)))
		X := bounds.Min.X + g.Gap
		for _, item := range spl {
			ret = append(ret, image.Rect(X, bounds.Min.Y+g.Gap, X+item, bounds.Max.Y-g.Gap))
			X += item + g.Gap
		}
	}
	return ret
}
