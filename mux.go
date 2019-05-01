package gui

import (
	"image"
	"image/draw"
	"sync"
)

type Mux struct {
	mu        sync.Mutex
	eventsIns []chan<- Event
	draw      chan<- func(draw.Image) image.Rectangle
}

func NewMux(env Env) (mux *Mux, master Env) {
	drawChan := make(chan func(draw.Image) image.Rectangle)
	mux = &Mux{draw: drawChan}
	master = mux.makeEnv(true)

	go func() {
		for d := range drawChan {
			env.Draw <- d
		}
		close(env.Draw)
	}()

	go func() {
		for e := range env.Events {
			mux.mu.Lock()
			for _, eventsIn := range mux.eventsIns {
				eventsIn <- e
			}
			mux.mu.Unlock()
		}
		mux.mu.Lock()
		for _, eventsIn := range mux.eventsIns {
			close(eventsIn)
		}
		mux.mu.Unlock()
	}()

	return mux, master
}

func (mux *Mux) MakeEnv() Env {
	return mux.makeEnv(false)
}

func (mux *Mux) makeEnv(master bool) Env {
	eventsOut, eventsIn := MakeEventsChan()
	drawChan := make(chan func(draw.Image) image.Rectangle)
	env := Env{eventsOut, drawChan}

	mux.mu.Lock()
	mux.eventsIns = append(mux.eventsIns, eventsIn)
	mux.mu.Unlock()

	go func() {
		for d := range drawChan {
			mux.draw <- d
		}
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
			i := -1
			for i = range mux.eventsIns {
				if mux.eventsIns[i] == eventsIn {
					break
				}
			}
			if i != -1 {
				mux.eventsIns = append(mux.eventsIns[:i], mux.eventsIns[i+1:]...)
			}
			mux.mu.Unlock()
		}
	}()

	return env
}
