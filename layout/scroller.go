package layout

import (
	"image"
	"image/color"
	"image/draw"
	// "log"
	"sync"

	"github.com/faiface/gui"
	"github.com/faiface/gui/win"
)

var _ Layout = &Scroller{}

type Scroller struct {
	Background  color.Color
	Length      int
	ChildHeight int
	Offset      int
	Gap         int
	Vertical    bool
}

func (s Scroller) redraw(drw draw.Image, bounds image.Rectangle) {
	col := s.Background
	if col == nil {
		col = image.Black
	}
	draw.Draw(drw, bounds, image.NewUniform(col), image.ZP, draw.Src)
}

func clamp(val, a, b int) int {
	if a > b {
		if val < b {
			return b
		}
		if val > a {
			return a
		}
	} else {
		if val > b {
			return b
		}
		if val < a {
			return a
		}
	}
	return val
}

func (s *Scroller) Intercept(env gui.Env) gui.Env {
	evs := env.Events()
	out, in := gui.MakeEventsChan()
	drawChan := make(chan func(draw.Image) image.Rectangle)
	ret := &muxEnv{out, drawChan}
	var lastResize gui.Resize
	var img draw.Image
	img = image.NewRGBA(image.ZR)
	var mu sync.Mutex
	var over bool

	go func() {
		for dc := range drawChan {
			mu.Lock()
			// draw.Draw will not draw out of bounds, call should be inexpensive if element not visible
			res := dc(img)
			// Only send a draw call up if visibly changed
			if res.Intersect(img.Bounds()) != image.ZR {
				env.Draw() <- func(drw draw.Image) image.Rectangle {
					draw.Draw(drw, lastResize.Rectangle, img, lastResize.Rectangle.Min, draw.Over)
					return img.Bounds()
				}
			}
			mu.Unlock()
		}
	}()

	go func() {
		for ev := range evs {
			switch ev := ev.(type) {
			case win.MoMove:
				mu.Lock()
				over = ev.Point.In(lastResize.Rectangle)
				mu.Unlock()
			case win.MoScroll:
				if !over {
					continue
				}
				mu.Lock()
				oldoff := s.Offset
				v := s.Length*s.ChildHeight + ((s.Length + 1) * s.Gap)
				if s.Vertical {
					h := lastResize.Dx()
					s.Offset = clamp(s.Offset+ev.Point.X*16, h-v, 0)
				} else {
					h := lastResize.Dy()
					s.Offset = clamp(s.Offset+ev.Point.Y*16, h-v, 0)
				}
				if oldoff != s.Offset {
					s.redraw(img, img.Bounds())
					in <- lastResize
				}
				mu.Unlock()
			case gui.Resize:
				mu.Lock()
				lastResize = ev
				img = image.NewRGBA(lastResize.Rectangle)
				s.redraw(img, img.Bounds())
				mu.Unlock()
				in <- ev
			default:
				in <- ev
			}
		}
	}()
	return ret
}

func (s Scroller) Lay(bounds image.Rectangle) []image.Rectangle {
	items := s.Length
	ch := s.ChildHeight
	gap := s.Gap

	ret := make([]image.Rectangle, items)
	Y := bounds.Min.Y + s.Offset + gap
	for i := 0; i < items; i++ {
		r := image.Rect(bounds.Min.X+gap, Y, bounds.Max.X-gap, Y+ch)
		ret[i] = r
		Y += ch + gap
	}
	return ret
}
