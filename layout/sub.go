package layout

import (
	"image"

	"github.com/faiface/gui/event"
)

func Sub(eif EventImageFlusher, r image.Rectangle) EventImageFlusher {
	s := &sub{
		eif: eif,
	}
	s.Event("resize", func(evt string) bool {
		var x1, y1, x2, y2 int
		event.Sscan(evt, &x1, &y1, &x2, &y2)
		r := image.Rect(x1, y1, x2, y2)
		s.sub = eif.Image().SubImage(r).(*image.RGBA)
		return false
	})
	s.Happen(event.Sprint("resize", r.Min.X, r.Min.Y, r.Max.X, r.Max.Y))
	return s
}

type sub struct {
	event.Dispatch
	eif EventImageFlusher
	sub *image.RGBA
}

func (s *sub) Image() *image.RGBA {
	return s.sub
}

func (s *sub) Flush(r image.Rectangle) {
	r = s.sub.Bounds().Intersect(r)
	s.eif.Flush(r)
}
