package layout

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/faiface/gui"
)

type box struct {
	Contents   []*gui.Env
	Background color.Color
	Split      SplitFunc
	Gap        int

	vertical bool
}

// NewBox creates a familiar flexbox-like list layout.
// It can be horizontal or vertical.
func NewBox(env gui.Env, contents []*gui.Env, options ...func(*box)) gui.Env {
	ret := &box{
		Background: image.Black,
		Contents:   contents,
		Split:      EvenSplit,
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

// BoxVertical changes the otherwise horizontal Box to be vertical.
func BoxVertical(b *box) {
	b.vertical = true
}

// BoxBackground changes the background of the box to a uniform color.
func BoxBackground(c color.Color) func(*box) {
	return func(grid *box) {
		grid.Background = c
	}
}

// BoxSplit changes the way the space is divided among the elements.
func BoxSplit(split SplitFunc) func(*box) {
	return func(grid *box) {
		grid.Split = split
	}
}

// BoxGap changes the box gap.
// The gap is identical everywhere (top, left, bottom, right).
func BoxGap(gap int) func(*box) {
	return func(grid *box) {
		grid.Gap = gap
	}
}

func (g *box) Redraw(drw draw.Image, bounds image.Rectangle) {
	draw.Draw(drw, bounds, image.NewUniform(g.Background), image.ZP, draw.Src)
}

func (g *box) Lay(bounds image.Rectangle) []image.Rectangle {
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
