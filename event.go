package gui

import "fmt"

// Event is a string encoding of an event. This may sound dangerous at first, but
// it enables nice pattern matching.
//
// Here are some examples of events (wi=window, mo=mouse, kb=keyboard):
//
//   wi/close
//   mo/down/421/890
//   kb/type/98
//   resize/920/655
//
// As you can see, the common way is to form the event string like a file path,
// from the most general information to the most specific. This allows pattern matching
// the prefix of the event, while ignoring the rest.
//
// Here's how to pattern match on an event:
//
//   switch {
//   case event.Matches("wi/close"):
//       // window closed
//   case event.Matches("mo/move/%d/%d", &x, &y):
//       // mouse moved to (x, y)
//       // mouse released on (x, y)
//   case event.Matches("kb/type/%d", &r):
//       // rune r typed on the keyboard (encoded as a number in the event string)
//   case event.Matches("resize/%d/%d", &w, &h):
//       // environment resized to (w, h)
//   }
//
// And here's how to pattern match on the prefix of an event:
//
//   switch {
//   case event.Matches("mo/"):
//       // this matches any mouse event
//   case event.Matches("kb/"):
//       // this matches any keyboard event
//   }
type Event string

// Eventf forms a new event. It works the same as fmt.Sprintf, except the return type is Event.
//
// For example:
//  Eventf("mo/down/%d/%d", x, y)
func Eventf(format string, a ...interface{}) Event {
	return Event(fmt.Sprintf(format, a...))
}

// Matches works the same as fmt.Sscanf, but returns a bool telling whether the match
// was successful. This makes it usable in a switch statement. See the doc for the Event
// type for an example.
func (e Event) Matches(format string, a ...interface{}) bool {
	_, err := fmt.Sscanf(string(e), format, a...)
	return err == nil
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
