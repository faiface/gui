package main

import (
	"image"
	"image/color"
	"image/draw"
	"os"
	"path/filepath"

	"github.com/faiface/gui"
	"golang.org/x/image/math/fixed"
)

func Browser(env gui.Env, theme *Theme, dir string, cd <-chan string, view chan<- string) {
	reload := func(dir string) (names []string, lineHeight int, namesImage *image.RGBA) {
		names = nil

		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if path == dir {
				return nil
			}
			rel, err := filepath.Rel(dir, path)
			if err != nil {
				return nil
			}
			if info.IsDir() {
				names = append(names, rel+string(filepath.Separator))
				return filepath.SkipDir
			}
			names = append(names, rel)
			return nil
		})

		var images []image.Image
		for _, name := range names {
			images = append(images, DrawText(name, theme.Face, theme.Text))
		}

		const inset = 4

		var width int
		for _, img := range images {
			if img.Bounds().Dx() > width {
				width = img.Bounds().Inset(-inset).Dx()
			}
		}

		metrics := theme.Face.Metrics()
		lineHeight = (metrics.Height + 2*fixed.I(inset)).Ceil()
		height := lineHeight * len(names)

		namesImage = image.NewRGBA(image.Rect(0, 0, width+2*inset, height+2*inset))
		for i := range images {
			r := image.Rect(
				0, lineHeight*i,
				width, lineHeight*(i+1),
			)
			DrawLeftCentered(namesImage, r.Inset(inset), images[i], draw.Over)
		}

		return names, lineHeight, namesImage
	}

	redraw := func(r image.Rectangle, selected int, position image.Point, lineHeight int, namesImage image.Image) func(draw.Image) image.Rectangle {
		return func(drw draw.Image) image.Rectangle {
			draw.Draw(drw, r, &image.Uniform{theme.Background}, image.ZP, draw.Src)
			draw.Draw(drw, r, namesImage, position, draw.Over)
			if selected >= 0 {
				highlightR := image.Rect(
					namesImage.Bounds().Min.X,
					namesImage.Bounds().Min.Y+lineHeight*selected,
					namesImage.Bounds().Max.X,
					namesImage.Bounds().Min.Y+lineHeight*(selected+1),
				)
				highlightR = highlightR.Sub(position).Add(r.Min)
				draw.DrawMask(
					drw, highlightR.Intersect(r),
					&image.Uniform{theme.Highlight}, image.ZP,
					&image.Uniform{color.Alpha{64}}, image.ZP,
					draw.Over,
				)
			}
			return r
		}
	}

	names, lineHeight, namesImage := reload(dir)

	var (
		r        image.Rectangle
		position = image.ZP
		selected = -1
	)

	for {
		select {
		case path := <-cd:
			if filepath.IsAbs(path) {
				dir = path
			} else {
				dir = filepath.Join(dir, path)
			}
			names, lineHeight, namesImage = reload(dir)
			position = image.ZP
			selected = -1
			env.Draw() <- redraw(r, selected, position, lineHeight, namesImage)

		case e, ok := <-env.Events():
			if !ok {
				close(env.Draw())
				return
			}

			var (
				x0, y0, x1, y1 int
				x, y                   int
			)

			switch {
			case e.Matches("resize/%d/%d/%d/%d", &x0, &y0, &x1, &y1):
				r = image.Rect(x0, y0, x1, y1)
				env.Draw() <- redraw(r, selected, position, lineHeight, namesImage)

			case e.Matches("mo/down/%d/%d", &x, &y):
				if !image.Pt(x, y).In(r) {
					continue
				}
				click := image.Pt(x, y).Sub(r.Min).Add(position)
				i := click.Y / lineHeight
				if i < 0 || i >= len(names) {
					continue
				}
				if selected == i {
					func() {
						path := filepath.Join(dir, names[selected])
						f, err := os.Open(path)
						if err != nil {
							return
						}
						defer f.Close()
						info, err := f.Stat()
						if err != nil {
							return
						}
						if info.IsDir() {
							dir = path
							names, lineHeight, namesImage = reload(dir)
							position = image.ZP
							selected = -1
							env.Draw() <- redraw(r, selected, position, lineHeight, namesImage)
						} else {
							view <- path
						}
					}()
				} else {
					selected = i
					env.Draw() <- redraw(r, selected, position, lineHeight, namesImage)
				}

			case e.Matches("mo/scroll/%d/%d", &x, &y):
				newP := position.Sub(image.Pt(int(x*16), int(y*16)))
				if newP.X > namesImage.Bounds().Max.X-r.Dx() {
					newP.X = namesImage.Bounds().Max.X - r.Dx()
				}
				if newP.Y > namesImage.Bounds().Max.Y-r.Dy() {
					newP.Y = namesImage.Bounds().Max.Y - r.Dy()
				}
				if newP.X < namesImage.Bounds().Min.X {
					newP.X = namesImage.Bounds().Min.X
				}
				if newP.Y < namesImage.Bounds().Min.Y {
					newP.Y = namesImage.Bounds().Min.Y
				}
				if newP != position {
					position = newP
					env.Draw() <- redraw(r, selected, position, lineHeight, namesImage)
				}
			}
		}
	}
}
