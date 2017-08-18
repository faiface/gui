package win

import (
	"image"
	"image/draw"
	"time"

	"github.com/faiface/gui/event"
	"github.com/faiface/mainthread"
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

func New(opts ...Option) (*Win, error) {
	o := options{
		title:     "",
		width:     640,
		height:    480,
		resizable: false,
	}
	for _, opt := range opts {
		opt(&o)
	}

	w := &Win{
		closed: make(chan struct{}),
	}

	var err error
	mainthread.Call(func() {
		w.w, err = makeGLFWWin(&o)
	})
	if err != nil {
		return nil, err
	}

	events := make(chan string)
	mainthread.Call(func() {
		w.resize(o.width, o.height)
		w.setUpEvents(events)
	})

	go func() {
		for {
			select {
			case event := <-events:
				w.Dispatch.Happen(event)
			case <-w.closed:
				return
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(time.Second / 120)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				mainthread.Call(glfw.PollEvents)
			case <-w.closed:
				return
			}
		}
	}()

	return w, nil
}

func makeGLFWWin(o *options) (*glfw.Window, error) {
	err := glfw.Init()
	if err != nil {
		return nil, err
	}
	glfw.WindowHint(glfw.DoubleBuffer, glfw.False)
	if o.resizable {
		glfw.WindowHint(glfw.Resizable, glfw.True)
	} else {
		glfw.WindowHint(glfw.Resizable, glfw.False)
	}
	w, err := glfw.CreateWindow(o.width, o.height, o.title, nil, nil)
	if err != nil {
		return nil, err
	}
	return w, nil
}

type Option func(*options)

func Title(title string) Option {
	return func(o *options) {
		o.title = title
	}
}

func Size(width, height int) Option {
	return func(o *options) {
		o.width = width
		o.height = height
	}
}

func Resizable() Option {
	return func(o *options) {
		o.resizable = true
	}
}

type options struct {
	title         string
	width, height int
	resizable     bool
}

type Win struct {
	event.Dispatch
	w      *glfw.Window
	rgba   *image.RGBA
	closed chan struct{}
}

func (w *Win) Close() error {
	return mainthread.CallErr(w.close)
}

func (w *Win) Image() *image.RGBA {
	return w.rgba
}

var curWin *Win = nil

func (w *Win) Flush(r image.Rectangle) {
	w.Dispatch.Happen(mkEvent("wi", "flush", r.Min.X, r.Min.Y, r.Max.X, r.Max.Y))
	mainthread.Call(func() {
		w.flush(r)
	})
}

func (w *Win) flush(r image.Rectangle) {
	if curWin != w {
		w.w.MakeContextCurrent()
		err := gl.Init()
		if err != nil {
			return
		}
		curWin = w
	}

	bounds := w.rgba.Bounds()
	r = bounds.Intersect(r)

	tmp := image.NewRGBA(r)
	draw.Draw(tmp, r, w.rgba, r.Min, draw.Src)

	gl.DrawBuffer(gl.FRONT)
	gl.Viewport(
		int32(bounds.Min.X),
		int32(bounds.Min.Y),
		int32(bounds.Dx()),
		int32(bounds.Dy()),
	)
	gl.RasterPos2d(
		-1+2*float64(r.Min.X)/float64(bounds.Dx()),
		+1-2*float64(r.Min.Y)/float64(bounds.Dy()),
	)
	gl.PixelZoom(1, -1)
	gl.DrawPixels(
		int32(r.Dx()),
		int32(r.Dy()),
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(tmp.Pix),
	)
	gl.Flush()
}

func (w *Win) close() error {
	close(w.closed)
	w.w.Destroy()
	return nil
}

func (w *Win) resize(width, height int) {
	bounds := image.Rect(0, 0, width, height)
	rgba := image.NewRGBA(bounds)
	if w.rgba != nil {
		draw.Draw(rgba, w.rgba.Bounds(), w.rgba, w.rgba.Bounds().Min, draw.Src)
	}
	w.rgba = rgba
	w.flush(bounds)
}
