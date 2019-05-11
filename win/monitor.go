package win

import (
  "github.com/faiface/mainthread"
  "github.com/go-gl/glfw/v3.2/glfw"
)


// Holds information on a monitor
type Monitor struct {
  Width, Height int
  RefreshRate   int
}

// Returns a struct containing information about the primary monitor
func GetPrimaryMonitor() (Monitor, error) {
  returns := mainthread.CallVal( func() interface{} {
    err := glfw.Init()
    if err != nil {
      return err
    }

    return glfw.GetPrimaryMonitor().GetVideoMode()
  })

  monitor := Monitor{0, 0, 0}
  var err error = nil

  switch v := returns.(type) {
    case *glfw.VidMode:
      videoMode := v

      width := videoMode.Width
      height := videoMode.Height
      refreshRate := videoMode.RefreshRate
      monitor = Monitor{width, height, refreshRate}
    case error:
      err = v
  }

  return monitor, err
}
