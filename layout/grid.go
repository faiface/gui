package layout

import (
	"image"
	"image/color"
	"image/draw"
)

type Grid struct {
	// Rows represents the number of childs of each row.
	Rows []int
	// Background represents the background of the grid as a uniform color.
	Background color.Color
	// Gap represents the grid gap, equal on all sides.
	Gap int
	// SplitX represents the way the space is divided among the columns in each row.
	SplitX SplitFunc
	// SplitY represents the way the space is divided among the rows.
	SplitY SplitFunc
}

func (g Grid) Redraw(drw draw.Image, bounds image.Rectangle) {
	col := g.Background
	if col == nil {
		col = image.Black
	}
	draw.Draw(drw, bounds, image.NewUniform(col), image.ZP, draw.Src)
}

func (g Grid) Lay(bounds image.Rectangle) []image.Rectangle {
	gap := g.Gap
	rows := g.Rows
	splitX := g.SplitX
	if splitX == nil {
		splitX = EvenSplit
	}
	splitY := g.SplitY
	if splitY == nil {
		splitY = EvenSplit
	}

	ret := make([]image.Rectangle, 0)
	rowsH := splitY(len(rows), bounds.Dy()-(gap*(len(rows)+1)))
	X := gap + bounds.Min.X
	Y := gap + bounds.Min.Y
	for y, cols := range rows {
		h := rowsH[y]
		colsW := splitX(cols, bounds.Dx()-(gap*(cols+1)))
		X = gap + bounds.Min.X
		for _, w := range colsW {
			ret = append(ret, image.Rect(X, Y, X+w, Y+h))
			X += gap + w
		}
		Y += gap + h
	}

	return ret
}
