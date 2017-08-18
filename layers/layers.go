package layers

import (
	"container/list"
	"errors"
	"image"
	"image/draw"
)

type Layers struct {
	dst    draw.Image
	r      image.Rectangle
	layers list.List
}

func (l *Layers) Dst(dst draw.Image, r image.Rectangle) {
	l.dst = dst
	l.r = r
	for e := l.layers.Front(); e != nil; e = e.Next() {
		layer := e.Value.(*Layer)
		rgba := image.NewRGBA(r)
		draw.Draw(rgba, layer.rgba.Bounds(), layer.rgba, layer.rgba.Bounds().Min, draw.Src)
		layer.rgba = rgba
	}
}

func (l *Layers) Push() *Layer {
	layer := &Layer{
		l:    l,
		rgba: image.NewRGBA(l.r),
	}
	layer.e = l.layers.PushBack(layer)
	return layer
}

func (l *Layers) Flush(r image.Rectangle) {
	if l.dst == nil {
		panic(errors.New("layers: Flush: no destination"))
	}
	draw.Draw(l.dst, r, image.Transparent, r.Min, draw.Src)
	for e := l.layers.Front(); e != nil; e = e.Next() {
		layer := e.Value.(*Layer)
		draw.Draw(l.dst, r, layer.rgba, r.Min, draw.Over)
	}
}

type Layer struct {
	l    *Layers
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
