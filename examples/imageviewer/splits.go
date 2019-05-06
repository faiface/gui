package main

import (
	"image"
	"image/draw"

	"github.com/faiface/gui"
)

type envPair struct {
	events <-chan gui.Event
	draw   chan<- func(draw.Image) image.Rectangle
}

func (ep *envPair) Events() <-chan gui.Event                      { return ep.events }
func (ep *envPair) Draw() chan<- func(draw.Image) image.Rectangle { return ep.draw }

func FixedLeft(env gui.Env, x1 int) gui.Env {
	out, in := gui.MakeEventsChan()

	go func() {
		for e := range env.Events() {
			var x0, y0, dummy, y1 int
			switch {
			case e.Matches("resize/%d/%d/%d/%d", &x0, &y0, &dummy, &y1):
				in <- gui.Eventf("resize/%d/%d/%d/%d", x0, y0, x1, y1)
			default:
				in <- e
			}
		}
		close(in)
	}()

	return &envPair{out, env.Draw()}
}

func FixedRight(env gui.Env, x0 int) gui.Env {
	out, in := gui.MakeEventsChan()

	go func() {
		for e := range env.Events() {
			var dummy, y0, x1, y1 int
			switch {
			case e.Matches("resize/%d/%d/%d/%d", &dummy, &y0, &x1, &y1):
				in <- gui.Eventf("resize/%d/%d/%d/%d", x0, y0, x1, y1)
			default:
				in <- e
			}
		}
		close(in)
	}()

	return &envPair{out, env.Draw()}
}

func FixedTop(env gui.Env, y1 int) gui.Env {
	out, in := gui.MakeEventsChan()

	go func() {
		for e := range env.Events() {
			var x0, y0, x1, dummy int
			switch {
			case e.Matches("resize/%d/%d/%d/%d", &x0, &y0, &x1, &dummy):
				in <- gui.Eventf("resize/%d/%d/%d/%d", x0, y0, x1, y1)
			default:
				in <- e
			}
		}
		close(in)
	}()

	return &envPair{out, env.Draw()}
}

func FixedBottom(env gui.Env, y0 int) gui.Env {
	out, in := gui.MakeEventsChan()

	go func() {
		for e := range env.Events() {
			var x0, dummy, x1, y1 int
			switch {
			case e.Matches("resize/%d/%d/%d/%d", &x0, &dummy, &x1, &y1):
				in <- gui.Eventf("resize/%d/%d/%d/%d", x0, y0, x1, y1)
			default:
				in <- e
			}
		}
		close(in)
	}()

	return &envPair{out, env.Draw()}
}

func EvenHorizontal(env gui.Env, minI, maxI, n int) gui.Env {
	out, in := gui.MakeEventsChan()

	go func() {
		for e := range env.Events() {
			var x0, y0, x1, y1 int
			switch {
			case e.Matches("resize/%d/%d/%d/%d", &x0, &y0, &x1, &y1):
				x0, x1 := x0+(x1-x0)*minI/n, x0+(x1-x0)*maxI/n
				in <- gui.Eventf("resize/%d/%d/%d/%d", x0, y0, x1, y1)
			default:
				in <- e
			}
		}
		close(in)
	}()

	return &envPair{out, env.Draw()}
}
