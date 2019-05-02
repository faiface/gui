package gui

import (
	"image"
	"image/draw"
)

type Env interface {
	Events() <-chan Event
	Draw() chan<- func(draw.Image) image.Rectangle
}
