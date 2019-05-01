package gui

import "fmt"

type Event string

func Eventf(format string, a ...interface{}) Event {
	return Event(fmt.Sprintf(format, a...))
}

func (e Event) Matches(format string, a ...interface{}) bool {
	_, err := fmt.Sscanf(string(e), format, a...)
	return err == nil
}

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
