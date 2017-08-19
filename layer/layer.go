package layers

import (
	"container/list"
	"errors"
	"image"
	"image/draw"

	"github.com/faiface/gui/event"
)

type ImageFlusher interface {
	Image() *image.RGBA
	Flush(r image.Rectangle)
}

type List struct {
	event.Dispatch
	dst    ImageFlusher
	layers list.List
}

func (l *List) Dst(dst ImageFlusher) {
	l.dst = dst
	for e := l.layers.Back(); e != nil; e = e.Prev() {
		layer := e.Value.(*Layer)
		rgba := image.NewRGBA(dst.Image().Bounds())
		draw.Draw(rgba, layer.rgba.Bounds(), layer.rgba, layer.rgba.Bounds().Min, draw.Src)
		layer.rgba = rgba
	}
}

func (l *List) Push() *Layer {
	layer := &Layer{
		l:    l,
		rgba: image.NewRGBA(l.dst.Image().Bounds()),
	}
	layer.e = l.layers.PushFront(layer)
	return layer
}

func (l *List) Flush(r image.Rectangle) {
	if l.dst == nil {
		panic(errors.New("layers: Flush: no destination"))
	}
	draw.Draw(l.dst.Image(), r, image.Transparent, r.Min, draw.Src)
	for e := l.layers.Back(); e != nil; e = e.Prev() {
		layer := e.Value.(*Layer)
		draw.Draw(l.dst.Image(), r, layer.rgba, r.Min, draw.Over)
	}
	l.dst.Flush(r)
}

func (l *List) Happen(event string) bool {
	l.Dispatch.Happen(event)
	for e := l.layers.Front(); e != nil; e = e.Next() {
		layer := e.Value.(*Layer)
		if layer.Happen(event) {
			return true
		}
	}
	return false
}

type Layer struct {
	event.Dispatch
	l    *List
	e    *list.Element
	rgba *image.RGBA
}

func (l *Layer) Remove() {
	l.l.layers.Remove(l.e)
}

func (l *Layer) Front() {
	l.l.layers.MoveToFront(l.e)
}

func (l *Layer) Image() *image.RGBA {
	return l.rgba
}

func (l *Layer) Flush(r image.Rectangle) {
	l.l.Flush(r)
}
