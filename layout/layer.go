package layout

import (
	"container/list"
	"errors"
	"image"
	"image/draw"

	"github.com/faiface/gui/event"
)

type EventImageFlusher interface {
	Event(pattern string, handler func(event string) bool)
	Image() *image.RGBA
	Flush(r image.Rectangle)
}

type LayerList struct {
	event.Dispatch
	dst    EventImageFlusher
	layers list.List
}

func NewLayerList(dst EventImageFlusher) *LayerList {
	l := &LayerList{dst: dst}

	dst.Event("", l.Happen)

	l.Event("resize", func(evt string) bool {
		var x1, y1, x2, y2 int
		event.Sscan(evt, &x1, &y1, &x2, &y2)

		for e := l.layers.Back(); e != nil; e = e.Prev() {
			layer := e.Value.(*Layer)
			rgba := image.NewRGBA(dst.Image().Bounds())
			draw.Draw(rgba, layer.rgba.Bounds(), layer.rgba, layer.rgba.Bounds().Min, draw.Src)
			layer.rgba = rgba
		}

		return false
	})

	r := dst.Image().Bounds()
	l.Happen(event.Sprint("resize", r.Min.X, r.Min.Y, r.Max.X, r.Max.Y))

	return l
}

func (l *LayerList) Add() *Layer {
	layer := &Layer{
		lst:  l,
		rgba: image.NewRGBA(l.dst.Image().Bounds()),
	}
	layer.elm = l.layers.PushFront(layer)
	return layer
}

func (l *LayerList) Remove(layer *Layer) {
	if layer.lst == nil {
		panic(errors.New("layer: Remove: layer already removed"))
	}
	l.layers.Remove(layer.elm)
	layer.lst = nil
}

func (l *LayerList) Front(layer *Layer) {
	if layer.lst == nil {
		panic(errors.New("layer: Front: layer removed"))
	}
	l.layers.MoveToFront(layer.elm)
}

func (l *LayerList) Happen(event string) bool {
	if l.Dispatch.Happen(event) {
		return true
	}
	for e := l.layers.Front(); e != nil; e = e.Next() {
		layer := e.Value.(*Layer)
		if layer.Happen(event) {
			return true
		}
	}
	return false
}

func (l *LayerList) Flush(r image.Rectangle) {
	if l.dst == nil {
		panic(errors.New("layer: Flush: no destination"))
	}
	draw.Draw(l.dst.Image(), r, image.Transparent, r.Min, draw.Src)
	for e := l.layers.Back(); e != nil; e = e.Prev() {
		layer := e.Value.(*Layer)
		draw.Draw(l.dst.Image(), r, layer.rgba, r.Min, draw.Over)
	}
	l.dst.Flush(r)
}

type Layer struct {
	event.Dispatch
	lst  *LayerList
	elm  *list.Element
	rgba *image.RGBA
}

func (l *Layer) List() *LayerList {
	return l.lst
}

func (l *Layer) Image() *image.RGBA {
	return l.rgba
}

func (l *Layer) Flush(r image.Rectangle) {
	if l.lst == nil {
		panic(errors.New("layer: Flush: layer removed"))
	}
	l.lst.Flush(r)
}
