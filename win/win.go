package win

import (
	"image"
	"image/draw"
	"runtime"
	"time"
	"unsafe"

	"github.com/faiface/gui"
	"github.com/faiface/mainthread"
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

// Option is a functional option to the window constructor New.
type Option func(*options)

type options struct {
	title         string
	width, height int
	resizable     bool
}

// Title option sets the title (caption) of the window.
func Title(title string) Option {
	return func(o *options) {
		o.title = title
	}
}

// Size option sets the width and height of the window.
func Size(width, height int) Option {
	return func(o *options) {
		o.width = width
		o.height = height
	}
}

// Resizable option makes the window resizable by the user.
func Resizable() Option {
	return func(o *options) {
		o.resizable = true
	}
}

// New creates a new window with all the supplied options.
//
// The default title is empty and the default size is 640x480.
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

	eventsOut, eventsIn := gui.MakeEventsChan()

	w := &Win{
		eventsOut: eventsOut,
		eventsIn:  eventsIn,
		draw:      make(chan func(draw.Image) image.Rectangle),
		newSize:   make(chan image.Rectangle),
		finish:    make(chan struct{}),
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

// Win is an Env that handles an actual graphical window.
//
// It receives its events from the OS and it draws to the surface of the window.
//
// Warning: only one window can be open at a time. This will be fixed.
//
// Here are all kinds of events that a window can produce, along with descriptions.
// Things enclosed in <> are values that are filled in.
//
//   resize/<w>/<h>            Window resized to w x h.
//   wi/close                  Window close button pressed.
//   mo/move/<x>/<y>           Mouse moved to (x, y).
//   mo/down/<x>/<y>/<button>  A mouse button pressed on (x, y).
//   mo/up/<x>/<y>/<button>    A mouse button released on (x, y).
//   kb/type/<code>            A unicode character typed on the keyboard.
//   kb/down/<key>             A key on the keyboard pressed.
//   kb/up/<key>               A key on the keyboard released.
//   kb/repeat/<key>           A key on the keyboard repeated (happens when held).
//
// <w>, <h>, <x>, <y>, and <code> are numbers (%d).
// <button> is one of:
//   left right middle
// <key> is one of:
//   left right up down escape space backspace delete enter
//   tab home end pageup pagedown shift ctrl alt
type Win struct {
	eventsOut <-chan gui.Event
	eventsIn  chan<- gui.Event
	draw      chan func(draw.Image) image.Rectangle

	newSize chan image.Rectangle
	finish  chan struct{}

	w   *glfw.Window
	img *image.RGBA
}

// Events returns the events channel of the window.
func (w *Win) Events() <-chan gui.Event { return w.eventsOut }

// Draw returns the draw channel of the window.
func (w *Win) Draw() chan<- func(draw.Image) image.Rectangle { return w.draw }

var buttonNames = map[glfw.MouseButton]string{
	glfw.MouseButtonLeft:   "left",
	glfw.MouseButtonRight:  "right",
	glfw.MouseButtonMiddle: "middle",
}

var keyNames = map[glfw.Key]string{
	glfw.KeyLeft:         "left",
	glfw.KeyRight:        "right",
	glfw.KeyUp:           "up",
	glfw.KeyDown:         "down",
	glfw.KeyEscape:       "escape",
	glfw.KeySpace:        "space",
	glfw.KeyBackspace:    "backspace",
	glfw.KeyDelete:       "delete",
	glfw.KeyEnter:        "enter",
	glfw.KeyTab:          "tab",
	glfw.KeyHome:         "home",
	glfw.KeyEnd:          "end",
	glfw.KeyPageUp:       "pageup",
	glfw.KeyPageDown:     "pagedown",
	glfw.KeyLeftShift:    "shift",
	glfw.KeyRightShift:   "shift",
	glfw.KeyLeftControl:  "ctrl",
	glfw.KeyRightControl: "ctrl",
	glfw.KeyLeftAlt:      "alt",
	glfw.KeyRightAlt:     "alt",
}

func (w *Win) eventThread() {
	var moX, moY int

	w.w.SetCursorPosCallback(func(_ *glfw.Window, x, y float64) {
		moX, moY = int(x), int(y)
		w.eventsIn <- gui.Eventf("mo/move/%d/%d", moX, moY)
	})

	w.w.SetMouseButtonCallback(func(_ *glfw.Window, button glfw.MouseButton, action glfw.Action, mod glfw.ModifierKey) {
		b, ok := buttonNames[button]
		if !ok {
			return
		}
		switch action {
		case glfw.Press:
			w.eventsIn <- gui.Eventf("mo/down/%d/%d/%s", moX, moY, b)
		case glfw.Release:
			w.eventsIn <- gui.Eventf("mo/up/%d/%d/%s", moX, moY, b)
		}
	})

	w.w.SetScrollCallback(func(_ *glfw.Window, xoff, yoff float64) {
		w.eventsIn <- gui.Eventf("mo/scroll/%d/%d", int(yoff), int(xoff))
	})

	w.w.SetCharCallback(func(_ *glfw.Window, r rune) {
		w.eventsIn <- gui.Eventf("kb/type/%d", r)
	})

	w.w.SetKeyCallback(func(_ *glfw.Window, key glfw.Key, _ int, action glfw.Action, _ glfw.ModifierKey) {
		k, ok := keyNames[key]
		if !ok {
			return
		}
		switch action {
		case glfw.Press:
			w.eventsIn <- gui.Eventf("kb/down/%s", k)
		case glfw.Release:
			w.eventsIn <- gui.Eventf("kb/up/%s", k)
		case glfw.Repeat:
			w.eventsIn <- gui.Eventf("kb/repeat/%s", k)
		}
	})

	w.w.SetSizeCallback(func(_ *glfw.Window, width, height int) {
		r := image.Rect(0, 0, width, height)
		w.newSize <- r
		w.eventsIn <- gui.Eventf("resize/%d/%d/%d/%d", r.Min.X, r.Min.Y, r.Max.X, r.Max.Y)
	})

	w.w.SetCloseCallback(func(_ *glfw.Window) {
		w.eventsIn <- gui.Eventf("wi/close")
	})

	r := w.img.Bounds()
	w.eventsIn <- gui.Eventf("resize/%d/%d/%d/%d", r.Min.X, r.Min.Y, r.Max.X, r.Max.Y)

	for {
		select {
		case <-w.finish:
			close(w.eventsIn)
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

	w.openGLFlush(w.img.Bounds())

loop:
	for {
		var totalR image.Rectangle

		select {
		case r := <-w.newSize:
			img := image.NewRGBA(r)
			draw.Draw(img, w.img.Bounds(), w.img, w.img.Bounds().Min, draw.Src)
			w.img = img
			totalR = totalR.Union(r)

		case d, ok := <-w.draw:
			if !ok {
				close(w.finish)
				return
			}
			r := d(w.img)
			totalR = totalR.Union(r)
		}

		for {
			select {
			case <-time.After(time.Second / 960):
				w.openGLFlush(totalR)
				totalR = image.ZR
				continue loop

			case r := <-w.newSize:
				img := image.NewRGBA(r)
				draw.Draw(img, w.img.Bounds(), w.img, w.img.Bounds().Min, draw.Src)
				w.img = img
				totalR = totalR.Union(r)

			case d, ok := <-w.draw:
				if !ok {
					close(w.finish)
					return
				}
				r := d(w.img)
				totalR = totalR.Union(r)
			}
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
		unsafe.Pointer(&tmp.Pix[0]),
	)
	gl.Flush()
}
