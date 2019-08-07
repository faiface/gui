package layout

import (
	"image"
	"image/color"
	"image/draw"
	"log"

	"github.com/faiface/gui"
)

var _ Layout = Grid{}

// Grid represents a grid with rows and columns in each row.
// Each row can be a different length.
type Grid struct {
	// Rows represents the number of childs of each row.
	Rows []int
	// Background represents the background of the grid as a uniform color.
	Background color.Color
	// Gap represents the grid gap, equal on all sides.
	Gap int
	// Split represents the way the space is divided among the columns in each row.
	Split SplitFunc
	// SplitRows represents the way the space is divided among the rows.
	SplitRows SplitFunc

	Margin      int
	Border      int
	BorderColor color.Color

	// Flip represents the orientation of the grid.
	// When false, rows are spread in the Y axis and columns in the X axis.
	// When true, rows are spread in the X axis and columns in the Y axis.
	Flip bool
}

func (g Grid) redraw(drw draw.Image, bounds image.Rectangle) {
	col := g.Background
	if col == nil {
		col = color.Black
	}
	if g.Border > 0 {
		bcol := g.BorderColor
		if bcol == nil {
			bcol = color.Black
		}
		draw.Draw(drw, bounds, image.NewUniform(bcol), image.ZP, draw.Src)
	}
	draw.Draw(drw, bounds.Inset(g.Border), image.NewUniform(col), image.ZP, draw.Src)
}

func (g Grid) Intercept(env gui.Env) gui.Env {
	return RedrawIntercepter{g.redraw}.Intercept(env)
}

func (g Grid) Lay(bounds image.Rectangle) []image.Rectangle {
	gap := g.Gap
	rows := g.Rows
	splitMain := g.Split
	if splitMain == nil {
		splitMain = EvenSplit
	}
	splitSec := g.SplitRows
	if splitSec == nil {
		splitSec = EvenSplit
	}
	margin := g.Margin
	flip := g.Flip
	if margin+gap < 0 {
		log.Println("Grid goes out of bounds")
	}
	if margin+gap < g.Border {
		log.Println("Grid border will not be shown properly")
	}

	ret := make([]image.Rectangle, 0)

	// Sorry it's not very understandable
	var H, W int
	var mX, mY int
	if flip {
		H = bounds.Dx()
		W = bounds.Dy()
		mX = bounds.Min.Y
		mY = bounds.Min.X
	} else {
		H = bounds.Dy()
		W = bounds.Dx()
		mX = bounds.Min.X
		mY = bounds.Min.Y
	}
	rowsH := splitSec(len(rows), H-(gap*(len(rows)+1))-margin*2)
	var X int
	var Y int
	Y = gap + mY + margin
	for y, cols := range rows {
		h := rowsH[y]
		colsW := splitMain(cols, W-(gap*(cols+1))-margin*2)
		X = gap + mX + margin
		for _, w := range colsW {
			var r image.Rectangle
			if flip {
				r = image.Rect(Y, X, Y+h, X+w)
			} else {
				r = image.Rect(X, Y, X+w, Y+h)
			}
			ret = append(ret, r)
			X += gap + w
		}
		Y += gap + h
	}

	return ret
}
