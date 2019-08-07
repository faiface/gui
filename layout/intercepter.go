package layout

import (
	"image"
	"image/draw"

	"github.com/faiface/gui"
)

// Intercepter represents an element that can interact with Envs.
// An Intercepter can modify Events, stop them or emit arbitrary ones.
// It can also put itself in the draw pipeline, for throttling very
// expensive draw calls for example.
type Intercepter interface {
	Intercept(gui.Env) gui.Env
}

var _ Intercepter = RedrawIntercepter{}

// RedrawIntercepter is a basic Intercepter, it is meant for use in simple Layouts
// that only need to redraw themselves.
type RedrawIntercepter struct {
	Redraw func(draw.Image, image.Rectangle)
}

// Intercept implements Intercepter
func (ri RedrawIntercepter) Intercept(env gui.Env) gui.Env {
	out, in := gui.MakeEventsChan()
	go func() {
		for e := range env.Events() {
			in <- e
			if resize, ok := e.(gui.Resize); ok {
				env.Draw() <- func(drw draw.Image) image.Rectangle {
					bounds := resize.Rectangle
					ri.Redraw(drw, bounds)
					return bounds
				}
			}
		}
	}()
	ret := &muxEnv{out, env.Draw()}
	return ret
}
