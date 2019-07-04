package layout

import (
	"image"
	"image/draw"
	"sync"

	"github.com/faiface/gui"
)

type Layout struct {
	masterEnv *MuxEnv
	inEvent   chan<- gui.Event

	mu         sync.Mutex
	lastResize gui.Event
	eventsIns  map[string]chan<- gui.Event
	draw       chan<- func(draw.Image) image.Rectangle

	Lay    func(image.Rectangle) map[string]image.Rectangle
	Redraw func(draw.Image, image.Rectangle)
}

func New(
	env gui.Env,
	lay func(image.Rectangle) map[string]image.Rectangle,
	redraw func(draw.Image, image.Rectangle),
) *Layout {

	mux := &Layout{
		Lay:    lay,
		Redraw: redraw,
	}
	drawChan := make(chan func(draw.Image) image.Rectangle)
	mux.draw = drawChan
	mux.masterEnv = mux.makeEnv("master", true)
	mux.inEvent = mux.masterEnv.In
	mux.eventsIns = make(map[string]chan<- gui.Event)
	go func() {
		for d := range drawChan {
			env.Draw() <- d
		}
		close(env.Draw())
	}()

	go func() {
		for e := range env.Events() {
			mux.inEvent <- e
		}
	}()

	go func() {
		for e := range mux.masterEnv.Events() {
			mux.mu.Lock()
			if resize, ok := e.(gui.Resize); ok {
				mux.lastResize = resize
				rect := resize.Rectangle

				mux.draw <- func(drw draw.Image) image.Rectangle {
					mux.Redraw(drw, rect)
					return rect
				}
				l := mux.Lay(rect)

				for key, eventsIn := range mux.eventsIns {
					func(rz gui.Resize) {
						rz.Rectangle = l[key]
						eventsIn <- rz
					}(resize)
				}
			} else {
				for _, eventsIn := range mux.eventsIns {
					eventsIn <- e
				}
			}
			mux.mu.Unlock()
		}
		mux.mu.Lock()
		for _, eventsIn := range mux.eventsIns {
			close(eventsIn)
		}
		mux.mu.Unlock()
	}()

	return mux
}

func (mux *Layout) GetEnv(name string) gui.Env {
	return mux.makeEnv(name, false)
}

type MuxEnv struct {
	In     chan<- gui.Event
	events <-chan gui.Event
	draw   chan<- func(draw.Image) image.Rectangle
}

func (m *MuxEnv) Events() <-chan gui.Event                      { return m.events }
func (m *MuxEnv) Draw() chan<- func(draw.Image) image.Rectangle { return m.draw }

// We do not store master env
func (mux *Layout) makeEnv(envName string, master bool) *MuxEnv {
	eventsOut, eventsIn := gui.MakeEventsChan()
	drawChan := make(chan func(draw.Image) image.Rectangle)
	env := &MuxEnv{eventsIn, eventsOut, drawChan}

	mux.mu.Lock()
	if !master {
		mux.eventsIns[envName] = eventsIn
	}

	// make sure to always send a resize event to a new Env if we got the size already
	// that means it missed the resize event by the root Env
	if mux.lastResize != nil {
		eventsIn <- mux.lastResize
	}
	mux.mu.Unlock()

	go func() {
		func() {
			// When the master Env gets its Draw() channel closed, it closes all the Events()
			// channels of all the children Envs, and it also closes the internal draw channel
			// of the Mux. Otherwise, closing the Draw() channel of the master Env wouldn't
			// close the Env the Mux is muxing. However, some child Envs of the Mux may still
			// send some drawing commmands before they realize that their Events() channel got
			// closed.
			//
			// That is perfectly fine if their drawing commands simply get ignored. This down here
			// is a little hacky, but (I hope) perfectly fine solution to the problem.
			//
			// When the internal draw channel of the Mux gets closed, the line marked with ! will
			// cause panic. We recover this panic, then we receive, but ignore all furhter draw
			// commands, correctly draining the Env until it closes itself.
			defer func() {
				if recover() != nil {
					for range drawChan {
					}
				}
			}()
			for d := range drawChan {
				mux.draw <- d // !
			}
		}()
		if master {
			mux.mu.Lock()
			for _, eventsIn := range mux.eventsIns {
				close(eventsIn)
			}
			mux.eventsIns = nil
			close(mux.draw)
			mux.mu.Unlock()
		} else {
			mux.mu.Lock()
			delete(mux.eventsIns, envName)

			close(eventsIn)
			mux.mu.Unlock()
		}
		if mux.lastResize != nil {
			mux.inEvent <- mux.lastResize
		}
	}()

	return env
}
