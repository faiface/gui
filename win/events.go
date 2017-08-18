package win

import (
	"fmt"
	"strings"

	"github.com/faiface/gui/event"
	"github.com/go-gl/glfw/v3.2/glfw"
)

func (w *Win) setUpEvents(events chan<- string) {
	var moX, moY int

	w.w.SetMouseButtonCallback(func(_ *glfw.Window, button glfw.MouseButton, action glfw.Action, mod glfw.ModifierKey) {
		switch action {
		case glfw.Press:
			sendEvent(events, "mo", "down", moX, moY)
		case glfw.Release:
			sendEvent(events, "mo", "up", moX, moY)
		}
	})

	w.w.SetCursorPosCallback(func(_ *glfw.Window, x, y float64) {
		moX, moY = int(x), int(y)
		sendEvent(events, "mo", "move", moX, moY)
	})

	w.w.SetCharCallback(func(_ *glfw.Window, r rune) {
		sendEvent(events, "kb", "type", r)
	})

	w.w.SetSizeCallback(func(_ *glfw.Window, width, height int) {
		w.resize(width, height)
		sendEvent(events, "wi", "resize", width, height)
	})

	w.w.SetCloseCallback(func(_ *glfw.Window) {
		events <- mkEvent("wi", "close")
		w.close()
	})
}

func sendEvent(events chan<- string, a ...interface{}) {
	go func() {
		events <- mkEvent(a...)
	}()
}

func mkEvent(a ...interface{}) string {
	s := make([]string, len(a))
	for i := range s {
		s[i] = fmt.Sprint(a[i])
	}
	return strings.Join(s, event.Sep)
}
