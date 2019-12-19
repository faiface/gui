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

// FrameDelay - usualy 1/60th of a second.
const FrameDelay time.Duration = time.Duration(int64(time.Second) / 60)

// Option is a functional option to the window constructor New.
type Option func(*options)

type options struct {
	title         string
	width, height int
	resizable     bool
	borderless    bool
	maximized     bool
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

// Borderless option makes the window borderless.
func Borderless() Option {
	return func(o *options) {
		o.borderless = true
	}
}

// Maximized option makes the window start maximized.
func Maximized() Option {
	return func(o *options) {
		o.maximized = true
	}
}

// New creates a new window with all the supplied options.
//
// The default title is empty and the default size is 640x480.
func New(opts ...Option) (*Win, error) {
	o := options{
		title:      "",
		width:      640,
		height:     480,
		resizable:  false,
		borderless: false,
		maximized:  false,
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

	mainthread.Call(func() {
		// hiDPI hack
		width, _ := w.w.GetFramebufferSize()
		w.ratio = width / o.width
		if w.ratio < 1 {
			w.ratio = 1
		}
		if w.ratio != 1 {
			o.width /= w.ratio
			o.height /= w.ratio
		}
		w.w.Destroy()
		w.w, err = makeGLFWWin(&o)
	})
	if err != nil {
		return nil, err
	}

	bounds := image.Rect(0, 0, o.width*w.ratio, o.height*w.ratio)
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
	if o.borderless {
		glfw.WindowHint(glfw.Decorated, glfw.False)
	}
	if o.maximized {
		glfw.WindowHint(glfw.Maximized, glfw.True)
	}
	w, err := glfw.CreateWindow(o.width, o.height, o.title, nil, nil)
	if err != nil {
		return nil, err
	}
	if o.maximized {
		o.width, o.height = w.GetFramebufferSize() // set o.width and o.height to the window size due to the window being maximized
	}
	return w, nil
}

// Win is an Env that handles an actual graphical window.
//
// It receives its events from the OS and it draws to the surface of the window.
//
// Warning: only one window can be open at a time. This will be fixed.
type Win struct {
	eventsOut <-chan gui.Event
	eventsIn  chan<- gui.Event
	draw      chan func(draw.Image) image.Rectangle

	newSize chan image.Rectangle
	finish  chan struct{}

	w     *glfw.Window
	img   *image.RGBA
	ratio int
}

// Events returns the events channel of the window.
func (w *Win) Events() <-chan gui.Event { return w.eventsOut }

// Draw returns the draw channel of the window.
func (w *Win) Draw() chan<- func(draw.Image) image.Rectangle { return w.draw }

var buttons = map[glfw.MouseButton]Button{
	glfw.MouseButtonLeft:   ButtonLeft,
	glfw.MouseButtonRight:  ButtonRight,
	glfw.MouseButtonMiddle: ButtonMiddle,
}

var keys = map[glfw.Key]Key{
	glfw.KeyLeft:         KeyLeft,
	glfw.KeyRight:        KeyRight,
	glfw.KeyUp:           KeyUp,
	glfw.KeyDown:         KeyDown,
	glfw.KeyEscape:       KeyEscape,
	glfw.KeySpace:        KeySpace,
	glfw.KeyBackspace:    KeyBackspace,
	glfw.KeyDelete:       KeyDelete,
	glfw.KeyEnter:        KeyEnter,
	glfw.KeyTab:          KeyTab,
	glfw.KeyHome:         KeyHome,
	glfw.KeyEnd:          KeyEnd,
	glfw.KeyPageUp:       KeyPageUp,
	glfw.KeyPageDown:     KeyPageDown,
	glfw.KeyLeftShift:    KeyShift,
	glfw.KeyRightShift:   KeyShift,
	glfw.KeyLeftControl:  KeyCtrl,
	glfw.KeyRightControl: KeyCtrl,
	glfw.KeyLeftAlt:      KeyAlt,
	glfw.KeyRightAlt:     KeyAlt,
}

func (w *Win) eventThread() {
	var moX, moY int

	w.w.SetCursorPosCallback(func(_ *glfw.Window, x, y float64) {
		moX, moY = int(x), int(y)
		w.eventsIn <- MoMove{image.Pt(moX*w.ratio, moY*w.ratio)}
	})

	w.w.SetMouseButtonCallback(func(_ *glfw.Window, button glfw.MouseButton, action glfw.Action, mod glfw.ModifierKey) {
		b, ok := buttons[button]
		if !ok {
			return
		}
		switch action {
		case glfw.Press:
			w.eventsIn <- MoDown{image.Pt(moX*w.ratio, moY*w.ratio), b}
		case glfw.Release:
			w.eventsIn <- MoUp{image.Pt(moX*w.ratio, moY*w.ratio), b}
		}
	})

	w.w.SetScrollCallback(func(_ *glfw.Window, xoff, yoff float64) {
		w.eventsIn <- MoScroll{image.Pt(int(xoff), int(yoff))}
	})

	w.w.SetCharCallback(func(_ *glfw.Window, r rune) {
		w.eventsIn <- KbType{r}
	})

	w.w.SetKeyCallback(func(_ *glfw.Window, key glfw.Key, _ int, action glfw.Action, _ glfw.ModifierKey) {
		k, ok := keys[key]
		if !ok {
			return
		}
		switch action {
		case glfw.Press:
			w.eventsIn <- KbDown{k}
		case glfw.Release:
			w.eventsIn <- KbUp{k}
		case glfw.Repeat:
			w.eventsIn <- KbRepeat{k}
		}
	})

	w.w.SetFramebufferSizeCallback(func(_ *glfw.Window, width, height int) {
		r := image.Rect(0, 0, width, height)
		w.newSize <- r
		w.eventsIn <- gui.Resize{Rectangle: r}
	})

	w.w.SetCloseCallback(func(_ *glfw.Window) {
		w.eventsIn <- WiClose{}
	})

	r := w.img.Bounds()
	w.eventsIn <- gui.Resize{Rectangle: r}

	t := time.Now()
	for {
		select {
		case <-w.finish:
			close(w.eventsIn)
			w.w.Destroy()
			return
		default:
			dif := time.Now().Sub(t)
			if dif > FrameDelay {
				glfw.WaitEventsTimeout(1.0 / 30)
				t = time.Now()
			}
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
			case <-time.After(FrameDelay):
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
