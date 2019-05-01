package gui

import (
	"image"
	"image/draw"
)

type Env struct {
	Events <-chan Event
	Draw   chan<- func(draw.Image) image.Rectangle
}
