# faiface/gui

Super minimal, rock-solid package for concurrent GUI in Go.

## Installation

```
go get -u github.com/faiface/gui
```

Currently uses [GLFW](https://www.glfw.org/) under the hood, so have [these dependencies](https://github.com/go-gl/glfw#installation).

## Overview

The idea of concurrent GUI pre-dates Go and is found in another language by Rob Pike called Newsqueak. He explains it quite nicely in [this talk](https://www.youtube.com/watch?v=hB05UFqOtFA&t=2408s). Newsqueak was similar to Go, mostly in that it had channels.

Why the hell has no one made a concurrent GUI in Go yet? I have no idea. Go is a perfect language for such a thing. Let's change that!

**This package is a minimal foundation for a concurrent GUI in Go.** I doesn't include widgets, layout systems, or anything like that. The main reason is that I am not yet sure how to do them most correctly in a concurrent GUI. So, instead of providing a half-assed, "fully-featured" library, I decided to make a small, rock-solid package, where everything is right.

**So, how does this work?**

There's [`Env`](https://godoc.org/github.com/faiface/gui#Env0), short for _environment_:

```go
type Env interface {
	Events() <-chan Event
	Draw() chan<- func(draw.Image) image.Rectangle
}
```

It's something that produces events (such as mouse clicks and key presses) and accepts draw commands.

Closing the `Draw()` channel destroys the environment. When destroyed (either by closing the `Draw()` channel or due to any other reason), the environment will always close the `Events()` channel.

As you can see, a draw command is a function that draws something onto a [`draw.Image`](https://golang.org/pkg/image/draw/#Image) and returns a rectangle telling which part got changed.

Yes, `faiface/gui` uses CPU for drawing. You won't make AAA games with it, but the performance is enough for most GUI apps. The benefits are outstanding, though:

1. Drawing is as simple as changing pixels.
2. No FPS (frames per second), results are immediately on the screen.
3. No need to organize the API around a GPU library, like OpenGL.
4. Use all the good packages, like [`"golang.org/x/image/font"`](https://godoc.org/golang.org/x/image/font) for fonts, or [`"github.com/fogleman/gg"`](https://godoc.org/github.com/fogleman/gg) for shapes.

What is an [`Event`](https://godoc.org/github.com/faiface/gui#Event)? It's a `string`:

```go
type Event string
```

Examples of `Event` strings are: `"wi/close"`, `"mo/move/104/320"`, `"kb/type/71"` (where `"wi"`, `"mo"`, and `"kb"` stand for _window_, _mouse_, and _keyboard_, respectively). The nice consequence of this form is that we can pattern match on it:

```go
switch {
case event.Matches("wi/close"):
    // window closed
case event.Matches("mo/move/%d/%d", &x, &y):
    // mouse moved to (x, y)
case event.Matches("mo/down/%d/%d", &x, &y):
    // mouse pressed on (x, y)
case event.Matches("mo/up/%d/%d", &x, &y):
    // mouse released on (x, y)
case event.Matches("kb/type/%d", &r):
    // rune r typed on the keyboard (encoded as a number in the event string)
case event.Matches("kb/down/%d", &k):
    // key k pressed on the keyboard
case event.Matches("kb/up/%d", &k):
    // key k released on the keyboard
case event.Matches("kb/repeat/%d", &k):
    // key k repeated on the keyboard (happens when held)
case event.Matches("resize/%d/%d", &w, &h):
    // environment resized to (w, h)
}
```

This shows all the possible events that a window can produce.

The current, limited list of key constants can be [found here](https://godoc.org/github.com/faiface/gui#pkg-constants).

The `"resize"` event is not prefixed with `"wi/"`, because it's not specific to windows. Events are easily extensible, for example, the `"mo/down/%d/%d"` event can be (and almost certainly will be) enriched with button of the mouse information, for example: `"mo/down/%d/%d/left"`.

How do we create a window? With the [`"github.com/faiface/gui/win"`](https://godoc.org/github.com/faiface/gui/win) package:

```go
// import "github.com/faiface/gui/win"
w, err := win.New(win.Title("faiface/win"), win.Size(800, 600), win.Resizable())
```

The [`win.New`](https://godoc.org/github.com/faiface/gui/win#New) constructor uses the [functional options pattern](https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis) by Dave Cheney. Unsurprisingly, the returned [`*win.Win`](https://godoc.org/github.com/faiface/gui/win#Win) is an `Env`.

Due to stupid limitations imposed by operating systems, the internal code that fetches events from the OS must run on the main thread of the program. To ensure this, we need to call [`mainthread.Run`](https://godoc.org/github.com/faiface/mainthread) in the `main` function:

```go
import "github.com/faiface/mainthread"

func run() {
    // do everything here
}

func main() {
    mainthread.Run(run)
}
```

How does it all look together? Here's a simple program that displays a nice, big rectangle in the middle of the window:

```go
package main

import (
	"image"
	"image/draw"

	"github.com/faiface/gui/win"
	"github.com/faiface/mainthread"
)

func run() {
	w, err := win.New(win.Title("faiface/gui"), win.Size(800, 600))
	if err != nil {
		panic(err)
	}

	w.Draw() <- func(drw draw.Image) image.Rectangle {
		r := image.Rect(200, 200, 600, 400)
		draw.Draw(drw, r, image.White, image.ZP, draw.Src)
		return r
	}

	for event := range w.Events() {
		switch {
		case event.Matches("wi/close"):
			close(w.Draw())
		}
	}
}

func main() {
	mainthread.Run(run)
}
```

### Muxing

When you receive an event from the `Events()` channel, it gets removed from the channel and no one else can receive it. But what if you have a button, a text field, four switches, and a bunch of other things that all want to receive the same events?

That's where multiplexing, or muxing comes in.

![Mux](images/mux.png)

A [`Mux`](https://godoc.org/github.com/faiface/gui#Mux) basically lets you split a single `Env` into multiple ones.

When the original `Env` produces an event, `Mux` sends it to each one of the multiple `Env`s.

When any one of the multiple `Env`s sends a draw function, `Mux` sends it to the original `Env`.

To mux an `Env`, use [`gui.NewMux`](https://godoc.org/github.com/faiface/gui#NewMux):

```go
mux, env := gui.NewMux(w)
```

Here we muxed the window `Env` stored in the `w` variable.

What's that second return value? That's the _master `Env`_. It's the first environment that the mux creates for us. It has a special role: if you close its `Draw()` channel, you close the `Mux`, all other `Env`s created by the `Mux`, and the original `Env`. But other than that, it's just like any other `Env` created by the `Mux`.

Don't use the original `Env` after muxing it. The `Mux` is using it and you'll steal its events at best.

To create more `Env`s, we can use [`mux.MakeEnv()`](https://godoc.org/github.com/faiface/gui#Mux.MakeEnv):

For example, here's a simple program that creates a blinking square whenever you click somewhere. The square will blink 3 times and then disappear:

```go
package main

import (
	"image"
	"image/draw"
	"time"

	"github.com/faiface/gui"
	"github.com/faiface/gui/win"
	"github.com/faiface/mainthread"
)

func Blinker(env gui.Env, r image.Rectangle) {
	for i := 0; i < 3; i++ {
		// switch to white
		env.Draw() <- func(drw draw.Image) image.Rectangle {
			draw.Draw(drw, r, image.White, image.ZP, draw.Src)
			return r
		}
		time.Sleep(time.Second / 2)

		// switch to black
		env.Draw() <- func(drw draw.Image) image.Rectangle {
			draw.Draw(drw, r, image.Black, image.ZP, draw.Src)
			return r
		}
		time.Sleep(time.Second / 2)
	}

	// close the environment
	close(env.Draw())
}

func run() {
	w, err := win.New(win.Title("faiface/gui"), win.Size(800, 600))
	if err != nil {
		panic(err)
	}

	mux, env := gui.NewMux(w)

	for event := range env.Events() {
		var x, y int
		switch {
		case event.Matches("wi/close"):
			close(env.Draw())

		case event.Matches("mo/down/%d/%d", &x, &y):
			go Blinker(mux.MakeEnv(), image.Rect(x-30, y-30, x+30, y+30))
		}
	}
}

func main() {
	mainthread.Run(run)
}
```

Closing the `Draw()` channel on an `Env` created by `mux.MakeEnv()` removes the `Env` from the `Mux`.

What if one of the `Env`s hangs and stops consuming events, or if it simply takes longer to consume them? Will all the other `Env`s hang as well?

They won't, because the channels of events have unlimited capacity and never block. This is implemented using an intermediate goroutine that handles the queueing.

![Events](images/events.png)

And that's basically all you need to know about `faiface/gui`! Happy hacking!

## What next?

This package is rock-solid, but nowhere near complete. Here are some of the things that I'd love to get done with your help:

- Get rid of the C dependencies.
- Mobile support.
- More event types and more key constants.
- A widgets/layout package.

Contributions are highly welcome!

## Licence

[MIT](LICENCE)
