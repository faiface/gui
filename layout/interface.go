package layout

import (
	"fmt"
	"image"
	"image/draw"
)

type EventDrawer interface {
	Event() <-chan EventConsume
	Draw() chan<- ImageFlush
}

type EventConsume struct {
	Event   string
	Consume chan<- bool
}

func SendEvent(ch chan<- EventConsume, format string, a ...interface{}) (consume <-chan bool) {
	cons := make(chan bool)
	ch <- EventConsume{fmt.Sprintf(format, a...), cons}
	return cons
}

func (ec EventConsume) Matches(format string, a ...interface{}) bool {
	_, err := fmt.Sscanf(ec.Event, format, a...)
	return err == nil
}

type ImageFlush struct {
	Image chan<- draw.Image
	Flush <-chan image.Rectangle
}

func SendDraw(ch chan<- ImageFlush) (img <-chan draw.Image, flush chan<- image.Rectangle) {
	imgC := make(chan draw.Image)
	flushC := make(chan image.Rectangle)
	ch <- ImageFlush{imgC, flushC}
	return imgC, flushC
}
