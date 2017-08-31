package layout

import (
	"fmt"
	"image"
	"image/draw"
)

type EventDrawer interface {
	Event() <-chan EventConsume
	Draw(func(draw.Image) image.Rectangle)
}

type EventConsume struct {
	Event
	Consume chan<- bool
}

type Event string

func (e Event) Matches(format string, a ...interface{}) bool {
	_, err := fmt.Sscanf(string(e), format, a...)
	return err == nil
}

func SendEvent(ch chan<- EventConsume, format string, a ...interface{}) (consume <-chan bool) {
	cons := make(chan bool)
	ch <- EventConsume{Event(fmt.Sprintf(format, a...)), cons}
	return cons
}
