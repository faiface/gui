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

func FixedLeft(env gui.Env, maxX int) gui.Env {
	out, in := gui.MakeEventsChan()

	go func() {
		for e := range env.Events() {
			var minX, minY, dummy, maxY int
			switch {
			case e.Matches("resize/%d/%d/%d/%d", &minX, &minY, &dummy, &maxY):
				in <- gui.Eventf("resize/%d/%d/%d/%d", minX, minY, maxX, maxY)
			default:
				in <- e
			}
		}
		close(in)
	}()

	return &envPair{out, env.Draw()}
}

func FixedRight(env gui.Env, minX int) gui.Env {
	out, in := gui.MakeEventsChan()

	go func() {
		for e := range env.Events() {
			var dummy, minY, maxX, maxY int
			switch {
			case e.Matches("resize/%d/%d/%d/%d", &dummy, &minY, &maxX, &maxY):
				in <- gui.Eventf("resize/%d/%d/%d/%d", minX, minY, maxX, maxY)
			default:
				in <- e
			}
		}
		close(in)
	}()

	return &envPair{out, env.Draw()}
}

func FixedTop(env gui.Env, maxY int) gui.Env {
	out, in := gui.MakeEventsChan()

	go func() {
		for e := range env.Events() {
			var minX, minY, maxX, dummy int
			switch {
			case e.Matches("resize/%d/%d/%d/%d", &minX, &minY, &maxX, &dummy):
				in <- gui.Eventf("resize/%d/%d/%d/%d", minX, minY, maxX, maxY)
			default:
				in <- e
			}
		}
		close(in)
	}()

	return &envPair{out, env.Draw()}
}

func FixedBottom(env gui.Env, minY int) gui.Env {
	out, in := gui.MakeEventsChan()

	go func() {
		for e := range env.Events() {
			var minX, dummy, maxX, maxY int
			switch {
			case e.Matches("resize/%d/%d/%d/%d", &minX, &dummy, &maxX, &maxY):
				in <- gui.Eventf("resize/%d/%d/%d/%d", minX, minY, maxX, maxY)
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
			var minX, minY, maxX, maxY int
			switch {
			case e.Matches("resize/%d/%d/%d/%d", &minX, &minY, &maxX, &maxY):
				minX, maxX := minX+(maxX-minX)*minI/n, minX+(maxX-minX)*maxI/n
				in <- gui.Eventf("resize/%d/%d/%d/%d", minX, minY, maxX, maxY)
			default:
				in <- e
			}
		}
		close(in)
	}()

	return &envPair{out, env.Draw()}
}
