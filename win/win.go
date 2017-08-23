package win

import (
	"image"
	"image/draw"
	"runtime"

	"github.com/faiface/gui/event"
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
		events: make(chan struct {
			event  string
			result chan<- bool
		}),
		flushes: make(chan struct {
			r       image.Rectangle
			flushed chan<- struct{}
		}),
	}

	var err error
	mainthread.Call(func() {
		w.w, err = makeGLFWWin(&o)
	})
	if err != nil {
		return nil, err
	}

	var (
		cancelOpenGLThread  = make(chan chan<- struct{})
		cancelEventDispatch = make(chan chan<- struct{})
		cancelEventThread   = make(chan chan<- struct{})
	)

	w.cancels = []chan<- chan<- struct{}{
		cancelEventThread,
		cancelEventDispatch,
		cancelOpenGLThread,
	}

	go func() {
		runtime.LockOSThread()
		openGLThread(w, cancelOpenGLThread, w.flushes)
	}()

	w.resize(o.width, o.height)

	go eventDispatch(w, cancelEventDispatch, w.events)
	mainthread.CallNonBlock(func() {
		eventThread(w, cancelEventThread)
	})

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
	dispatch event.Dispatch
	w        *glfw.Window
	rgba     *image.RGBA
	events   chan struct {
		event  string
		result chan<- bool
	}
	flushes chan struct {
		r       image.Rectangle
		flushed chan<- struct{}
	}
	cancels []chan<- chan<- struct{}
}

func (w *Win) Event(pattern string, handler func(evt string) bool) {
	w.dispatch.Event(pattern, handler)
}

func (w *Win) Happen(evt string) bool {
	result := make(chan bool)
	w.events <- struct {
		event  string
		result chan<- bool
	}{evt, result}
	return <-result
}

func (w *Win) Close() error {
	go func() {
		for _, cancel := range w.cancels {
			confirm := make(chan struct{})
			cancel <- confirm
			<-confirm
		}
	}()
	return nil
}

func (w *Win) Image() *image.RGBA {
	return w.rgba
}

func (w *Win) Flush(r image.Rectangle) {
	flushed := make(chan struct{})
	w.flushes <- struct {
		r       image.Rectangle
		flushed chan<- struct{}
	}{r, flushed}
	<-flushed
}

func (w *Win) resize(width, height int) {
	bounds := image.Rect(0, 0, width, height)
	rgba := image.NewRGBA(bounds)
	if w.rgba != nil {
		draw.Draw(rgba, w.rgba.Bounds(), w.rgba, w.rgba.Bounds().Min, draw.Src)
	}
	w.rgba = rgba
	w.Flush(bounds)
}

func eventDispatch(w *Win, cancel <-chan chan<- struct{}, events chan struct {
	event  string
	result chan<- bool
}) {
loop:
	for {
		select {
		case evt := <-events:
			evt.result <- w.dispatch.Happen(evt.event)
		case confirm := <-cancel:
			close(confirm)
			break loop
		}
	}
}

func eventThread(w *Win, cancel <-chan chan<- struct{}) {
	var moX, moY int

	w.w.SetMouseButtonCallback(func(_ *glfw.Window, button glfw.MouseButton, action glfw.Action, mod glfw.ModifierKey) {
		switch action {
		case glfw.Press:
			w.Happen(event.Sprint("mo", "down", moX, moY))
		case glfw.Release:
			w.Happen(event.Sprint("mo", "up", moX, moY))
		}
	})

	w.w.SetCursorPosCallback(func(_ *glfw.Window, x, y float64) {
		moX, moY = int(x), int(y)
		w.Happen(event.Sprint("mo", "move", moX, moY))
	})

	w.w.SetCharCallback(func(_ *glfw.Window, r rune) {
		w.Happen(event.Sprint("kb", "type", r))
	})

	w.w.SetSizeCallback(func(_ *glfw.Window, width, height int) {
		w.resize(width, height)
		w.Happen(event.Sprint("resize", 0, 0, width, height))
	})

	w.w.SetCloseCallback(func(_ *glfw.Window) {
		w.Happen(event.Sprint("wi", "close"))
	})

loop:
	for {
		select {
		default:
			glfw.WaitEvents()
		case confirm := <-cancel:
			close(confirm)
			break loop
		}
	}
}

func openGLThread(w *Win, cancel <-chan chan<- struct{}, flushes <-chan struct {
	r       image.Rectangle
	flushed chan<- struct{}
}) {
	w.w.MakeContextCurrent()
	gl.Init()

loop:
	for {
		select {
		case flush := <-w.flushes:
			r := flush.r

			bounds := w.rgba.Bounds()
			r = bounds.Intersect(r)
			if r.Empty() {
				return
			}

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

			close(flush.flushed)
		case confirm := <-cancel:
			w.w.Destroy()
			close(confirm)
			break loop
		}
	}
}
