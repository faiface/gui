package layout

import (
	"fmt"
	"image"
	"image/draw"
)

type EventDrawer interface {
	Event() <-chan Event
	Draw() chan<- func(draw.Image) image.Rectangle
}

type Event string

func Eventf(format string, a ...interface{}) Event {
	return Event(fmt.Sprintf(format, a...))
}

func (e Event) Matches(format string, a ...interface{}) bool {
	_, err := fmt.Sscanf(string(e), format, a...)
	return err == nil
}
