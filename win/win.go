package win

import (
	"image"
	"image/draw"
	"runtime"

	"github.com/faiface/gui/layout"
	"github.com/faiface/mainthread"
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

type Option func(*options)

type options struct {
	title         string
	width, height int
	resizable     bool
}

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
		event:   make(chan layout.Event),
		draw:    make(chan func(draw.Image) image.Rectangle),
		newSize: make(chan image.Rectangle),
		finish:  make(chan struct{}),
	}

	var err error
	mainthread.Call(func() {
		w.w, err = makeGLFWWin(&o)
	})
	if err != nil {
		return nil, err
	}

	bounds := image.Rect(0, 0, o.width, o.height)
	w.img = image.NewRGBA(bounds)

	go func() {
		runtime.LockOSThread()
		w.openGLThread()
	}()

	mainthread.CallNonBlock(w.eventThread)

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

type Win struct {
	event chan layout.Event
	draw  chan func(draw.Image) image.Rectangle

	newSize chan image.Rectangle
	finish  chan struct{}

	w   *glfw.Window
	img *image.RGBA
}

func (w *Win) Event() <-chan layout.Event {
	return w.event
}

func (w *Win) Draw() chan<- func(draw.Image) image.Rectangle {
	return w.draw
}

func (w *Win) eventThread() {
	var moX, moY int

	w.w.SetMouseButtonCallback(func(_ *glfw.Window, button glfw.MouseButton, action glfw.Action, mod glfw.ModifierKey) {
		switch action {
		case glfw.Press:
			w.event <- layout.Eventf("mo/down/%d/%d", moX, moY)
		case glfw.Release:
			w.event <- layout.Eventf("mo/up/%d/%d", moX, moY)
		}
	})

	w.w.SetCursorPosCallback(func(_ *glfw.Window, x, y float64) {
		moX, moY = int(x), int(y)
		w.event <- layout.Eventf("mo/move/%d/%d", moX, moY)
	})

	w.w.SetCharCallback(func(_ *glfw.Window, r rune) {
		w.event <- layout.Eventf("kb/type/%d", r)
	})

	w.w.SetSizeCallback(func(_ *glfw.Window, width, height int) {
		r := image.Rect(0, 0, width, height)
		w.newSize <- r
		w.event <- layout.Eventf("resize/%d/%d/%d/%d", r.Min.X, r.Min.Y, r.Max.X, r.Max.Y)
	})

	w.w.SetCloseCallback(func(_ *glfw.Window) {
		w.event <- layout.Eventf("wi/close")
	})

	r := w.img.Bounds()
	w.event <- layout.Eventf("resize/%d/%d/%d/%d", r.Min.X, r.Min.Y, r.Max.X, r.Max.Y)

	for {
		select {
		case <-w.finish:
			close(w.event)
			w.w.Destroy()
			return
		default:
			glfw.WaitEventsTimeout(1.0 / 30)
		}
	}
}

func (w *Win) openGLThread() {
	w.w.MakeContextCurrent()
	gl.Init()

	for {
		select {
		case r := <-w.newSize:
			img := image.NewRGBA(r)
			draw.Draw(img, w.img.Bounds(), w.img, w.img.Bounds().Min, draw.Src)
			w.img = img
			w.openGLFlush(r)

		case d, ok := <-w.draw:
			if !ok {
				close(w.finish)
				return
			}
			r := d(w.img)
			w.openGLFlush(r)
		}
	}
}

func (w *Win) openGLFlush(r image.Rectangle) {
	bounds := w.img.Bounds()
	r = r.Intersect(bounds)
	if r.Empty() {
		return
	}

	tmp := image.NewRGBA(r)
	draw.Draw(tmp, r, w.img, r.Min, draw.Src)

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
