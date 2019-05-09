package gui

import (
	"fmt"
	"image"
)

// Event is something that can happen in an environment.
//
// This package defines only one kind of event: Resize. Other packages implementing environments
// may implement more kinds of events. For example, the win package implements all kinds of
// events for mouse and keyboard.
type Event interface {
	String() string
}

// Resize is an event that happens when the environment changes the size of its drawing area.
type Resize struct {
	image.Rectangle
}

func (r Resize) String() string {
	return fmt.Sprintf("resize/%d/%d/%d/%d", r.Min.X, r.Min.Y, r.Max.X, r.Max.Y)
}

// MakeEventsChan implements a channel of events with an unlimited capacity. It does so
// by creating a goroutine that queues incoming events. Sending to this channel never blocks
// and no events get lost.
//
// The unlimited capacity channel is very suitable for delivering events because the consumer
// may be unavailable for some time (doing a heavy computation), but will get to the events
// later.
//
// An unlimited capacity channel has its dangers in general, but is completely fine for
// the purpose of delivering events. This is because the production of events is fairly
// infrequent and should never out-run their consumption in the long term.
func MakeEventsChan() (<-chan Event, chan<- Event) {
	out, in := make(chan Event), make(chan Event)

	go func() {
		var queue []Event

		for {
			x, ok := <-in
			if !ok {
				close(out)
				return
			}
			queue = append(queue, x)

			for len(queue) > 0 {
				select {
				case out <- queue[0]:
					queue = queue[1:]
				case x, ok := <-in:
					if !ok {
						for _, x := range queue {
							out <- x
						}
						close(out)
						return
					}
					queue = append(queue, x)
				}
			}
		}
	}()

	return out, in
}
