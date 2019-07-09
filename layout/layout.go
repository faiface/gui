package layout

import (
	"image"
	"image/draw"
)

// Layout represents any graphical layout
//
// A Layout needs to be able to redraw itself with the Redraw method.
// Redraw() only draws the background or frame of the Layout, not the childs.
//
// Lay represents the way to divide space among your childs.
// It takes a parameter of how much space is available,
// and returns where exactly to put its childs.
type Layout interface {
	Lay(image.Rectangle) []image.Rectangle
	Redraw(draw.Image, image.Rectangle)
}
