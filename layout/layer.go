package layout

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

type LayerList struct {
	event.Dispatch
	dst    ImageFlusher
	layers list.List
}

func (l *LayerList) Dst(dst ImageFlusher) {
	l.dst = dst
	for e := l.layers.Back(); e != nil; e = e.Prev() {
		layer := e.Value.(*Layer)
		rgba := image.NewRGBA(dst.Image().Bounds())
		draw.Draw(rgba, layer.rgba.Bounds(), layer.rgba, layer.rgba.Bounds().Min, draw.Src)
		layer.rgba = rgba
	}
}

func (l *LayerList) Push() *Layer {
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
