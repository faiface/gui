# Paint

This is a simple paint program. Pick colors on the right, draw on the rest.

![Screenshot](screenshot.png)

The canvas in the middle is controlled by a goroutine. Each of the 8 color pickers is also controlled by its own goroutine. When clicked, the color picker sends its color to the canvas over a channel. The canvas listens on the channel and changes its color.

Uses [`fogleman/gg`](https://github.com/fogleman/gg) for drawing lines.
