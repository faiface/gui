package layout

import (
	"image"
)

// Layout represents any graphical layout
//
// Lay represents the way to divide space among your childs.
// It takes a parameter of how much space is available,
// and returns where exactly to put its childs.
//
// Intercept transforms an Env channel to another.
// This way the Layout can emit its own Events, re-emit previous ones,
// or even stop an event from propagating, think win.MoScroll.
// It can be a no-op.
type Layout interface {
	Lay(image.Rectangle) []image.Rectangle
	Intercepter
}
