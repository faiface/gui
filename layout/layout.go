package layout

import (
	"image"
	"image/draw"
)

// Layout represents any graphical layout
//
// Lay represents the way to divide space among your childs.
// It takes a parameter of how much space is available,
// and returns where exactly to put its childs.
// The order must be the same as Items.
//
// Redraw only draws the background or frame of the Layout, not the childs.
type Layout interface {
	Lay(image.Rectangle) []image.Rectangle
	Redraw(draw.Image, image.Rectangle)
}
