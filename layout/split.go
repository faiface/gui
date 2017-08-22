package layout

import (
	"image"

	"github.com/faiface/gui/event"
)

type Box struct {
	event.Dispatch
	dst EventImageFlusher
	sub *image.RGBA
}

func (b *Box) Image() *image.RGBA {
	return b.sub
}

func (b *Box) Flush(r image.Rectangle) {
	r = r.Intersect(b.dst.Image().Bounds())
	b.dst.Flush(r)
}

type Splitter func(r image.Rectangle, i, n int) image.Rectangle

func Vertical() Splitter {
	return func(r image.Rectangle, i, n int) image.Rectangle {
		width := r.Dx()
		return image.Rect(
			r.Min.X+width*i/n,
			r.Min.Y,
			r.Min.X+width*(i+1)/n,
			r.Max.Y,
		)
	}
}

func Horizontal() Splitter {
	return func(r image.Rectangle, i, n int) image.Rectangle {
		height := r.Dy()
		return image.Rect(
			r.Min.X,
			r.Min.Y+height*i/n,
			r.Max.X,
			r.Min.Y+height*(i+1)/n,
		)
	}
}

func FixedTop(thickness int, rest Splitter) Splitter {
	return VariableTop(&thickness, rest)
}

func FixedBottom(thickness int, rest Splitter) Splitter {
	return VariableBottom(&thickness, rest)
}

func FixedLeft(thickness int, rest Splitter) Splitter {
	return VariableLeft(&thickness, rest)
}

func FixedRight(thickness int, rest Splitter) Splitter {
	return VariableRight(&thickness, rest)
}

func VariableTop(thickness *int, rest Splitter) Splitter {
	return func(r image.Rectangle, i, n int) image.Rectangle {
		if i == 0 {
			r.Max.Y = r.Min.Y + *thickness
			return r
		}
		r.Min.Y += *thickness
		return rest(r, i-1, n-1)
	}
}

func VariableBottom(thickness *int, rest Splitter) Splitter {
	return func(r image.Rectangle, i, n int) image.Rectangle {
		if i == 0 {
			r.Min.Y = r.Max.Y - *thickness
			return r
		}
		r.Max.Y -= *thickness
		return rest(r, i-1, n-1)
	}
}

func VariableLeft(thickness *int, rest Splitter) Splitter {
	return func(r image.Rectangle, i, n int) image.Rectangle {
		if i == 0 {
			r.Max.X = r.Min.X + *thickness
			return r
		}
		r.Min.X += *thickness
		return rest(r, i-1, n-1)
	}
}

func VariableRight(thickness *int, rest Splitter) Splitter {
	return func(r image.Rectangle, i, n int) image.Rectangle {
		if i == 0 {
			r.Min.X = r.Max.X - *thickness
			return r
		}
		r.Max.X -= *thickness
		return rest(r, i-1, n-1)
	}
}

type Split struct {
	event.Dispatch
	dst      EventImageFlusher
	splitter Splitter
	boxes    []*Box
}

func NewSplit(dst EventImageFlusher, splitter Splitter) *Split {
	s := &Split{
		dst:      dst,
		splitter: splitter,
	}
	dst.Event("", s.Happen)
	s.Event("resize", func(evt string) bool {
		s.Split()
		return true
	})
	return s
}

func (s *Split) Happen(evt string) bool {
	if s.Dispatch.Happen(evt) {
		return true
	}
	for _, box := range s.boxes {
		if box.Happen(evt) {
			return true
		}
	}
	return false
}

func (s *Split) Add() *Box {
	box := &Box{
		dst: s.dst,
	}
	s.boxes = append(s.boxes, box)
	return box
}

func (s *Split) Split() {
	r := s.dst.Image().Bounds()
	n := len(s.boxes)
	for i := range s.boxes {
		subR := s.splitter(r, i, n)
		s.boxes[i].sub = s.dst.Image().SubImage(subR).(*image.RGBA)
	}
	for i := range s.boxes {
		subR := s.boxes[i].sub.Bounds()
		s.boxes[i].Happen(event.Sprint("resize", subR.Min.X, subR.Min.Y, subR.Max.X, subR.Max.Y))
	}
}
