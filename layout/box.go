package layout

import (
	"image"
	"image/color"
	"image/draw"
)

type Box struct {
	// Number of child elements
	Length int
	// Background changes the background of the Box to a uniform color.
	Background color.Color
	// Split changes the way the space is divided among the elements.
	Split SplitFunc
	// Gap changes the Box gap.
	// The gap is identical everywhere (top, left, bottom, right).
	Gap int

	// Vertical changes the otherwise horizontal Box to be vertical.
	Vertical bool
}

func (b Box) Redraw(drw draw.Image, bounds image.Rectangle) {
	col := b.Background
	if col == nil {
		col = image.Black
	}

	draw.Draw(drw, bounds, image.NewUniform(col), image.ZP, draw.Src)
}

func (b Box) Lay(bounds image.Rectangle) []image.Rectangle {
	items := b.Length
	gap := b.Gap
	split := b.Split
	if split == nil {
		split = EvenSplit
	}
	ret := make([]image.Rectangle, 0, items)
	if b.Vertical {
		spl := split(items, bounds.Dy()-(gap*(items+1)))
		Y := bounds.Min.Y + gap
		for _, item := range spl {
			ret = append(ret, image.Rect(bounds.Min.X+gap, Y, bounds.Max.X-gap, Y+item))
			Y += item + gap
		}
	} else {
		spl := split(items, bounds.Dx()-(gap*(items+1)))
		X := bounds.Min.X + gap
		for _, item := range spl {
			ret = append(ret, image.Rect(X, bounds.Min.Y+gap, X+item, bounds.Max.Y-gap))
			X += item + gap
		}
	}
	return ret
}
