package main

import (
	"image/color"

	"golang.org/x/image/font"
)

type Theme struct {
	Face font.Face

	Background color.Color
	Empty      color.Color
	Text       color.Color
	Highlight  color.Color
	ButtonUp   color.Color
	ButtonOver color.Color
	ButtonDown color.Color
}
