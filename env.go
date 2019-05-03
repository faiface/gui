package gui

import (
	"image"
	"image/draw"
)

// Env is the most important thing in this package. It is an interactive graphical
// environment, such as a window.
//
// It has two channels: Events() and Draw().
//
// The Events() channel produces events, like mouse and keyboard presses, while the
// Draw() channel receives drawing functions. A drawing function draws onto the
// supplied draw.Image, which is the drawing area of the Env and returns a rectangle
// covering the whole part of the image that got changed.
//
// The Events() channel must be unlimited in capacity. Use MakeEventsChan() to create
// a channel of events with an unlimited capacity.
//
// The Draw() channel may be synchronous.
//
// Drawing functions sent to the Draw() channel are not guaranteed to be executed.
//
// Closing the Draw() channel results in closing the Env. The Env will subsequently
// close the Events() channel. On the other hand, when the Events() channel gets closed
// the user of the Env should subsequently close the Draw() channel.
type Env interface {
	Events() <-chan Event
	Draw() chan<- func(draw.Image) image.Rectangle
}
