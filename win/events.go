package win

import (
	"fmt"
	"image"
)

// Button indicates a mouse button in an event.
type Button string

// List of all mouse buttons.
const (
	ButtonLeft   Button = "left"
	ButtonRight  Button = "right"
	ButtonMiddle Button = "middle"
)

// Key indicates a keyboard key in an event.
type Key string

// List of all keyboard keys.
const (
	KeyLeft      Key = "left"
	KeyRight     Key = "right"
	KeyUp        Key = "up"
	KeyDown      Key = "down"
	KeyEscape    Key = "escape"
	KeySpace     Key = "space"
	KeyBackspace Key = "backspace"
	KeyDelete    Key = "delete"
	KeyEnter     Key = "enter"
	KeyTab       Key = "tab"
	KeyHome      Key = "home"
	KeyEnd       Key = "end"
	KeyPageUp    Key = "pageup"
	KeyPageDown  Key = "pagedown"
	KeyShift     Key = "shift"
	KeyCtrl      Key = "ctrl"
	KeyAlt       Key = "alt"
)

type (
	// WiClose is an event that happens when the user presses the close button on the window.
	WiClose struct{}

	// WiRefresh is an event that happens when the user opens the window after minimizing.
	WiRefresh struct{}

	// MoMove is an event that happens when the mouse gets moved across the window.
	MoMove struct{ image.Point }

	// MoDown is an event that happens when a mouse button gets pressed.
	MoDown struct {
		image.Point
		Button Button
	}

	// MoUp is an event that happens when a mouse button gets released.
	MoUp struct {
		image.Point
		Button Button
	}

	// MoScroll is an event that happens on scrolling the mouse.
	//
	// The Point field tells the amount scrolled in each direction.
	MoScroll struct{ image.Point }

	// KbType is an event that happens when a Unicode character gets typed on the keyboard.
	KbType struct{ Rune rune }

	// KbDown is an event that happens when a key on the keyboard gets pressed.
	KbDown struct{ Key Key }

	// KbUp is an event that happens when a key on the keyboard gets released.
	KbUp struct{ Key Key }

	// KbRepeat is an event that happens when a key on the keyboard gets repeated.
	//
	// This happens when its held down for some time.
	KbRepeat struct{ Key Key }
)

func (WiClose) String() string     { return "wi/close" }
func (WiRefresh) String() string   { return "wi/refresh" }
func (mm MoMove) String() string   { return fmt.Sprintf("mo/move/%d/%d", mm.X, mm.Y) }
func (md MoDown) String() string   { return fmt.Sprintf("mo/down/%d/%d/%s", md.X, md.Y, md.Button) }
func (mu MoUp) String() string     { return fmt.Sprintf("mo/up/%d/%d/%s", mu.X, mu.Y, mu.Button) }
func (ms MoScroll) String() string { return fmt.Sprintf("mo/scroll/%d/%d", ms.X, ms.Y) }
func (kt KbType) String() string   { return fmt.Sprintf("kb/type/%d", kt.Rune) }
func (kd KbDown) String() string   { return fmt.Sprintf("kb/down/%s", kd.Key) }
func (ku KbUp) String() string     { return fmt.Sprintf("kb/up/%s", ku.Key) }
func (kr KbRepeat) String() string { return fmt.Sprintf("kb/repeat/%s", kr.Key) }
