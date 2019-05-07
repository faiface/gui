package main

import (
	"image"
	"image/color"
	"image/draw"
	"sync"

	"github.com/golang/freetype/truetype"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

type concurrentFace struct {
	mu   sync.Mutex
	face font.Face
}

func (cf *concurrentFace) Close() error {
	cf.mu.Lock()
	defer cf.mu.Unlock()
	return cf.face.Close()
}

func (cf *concurrentFace) Glyph(dot fixed.Point26_6, r rune) (dr image.Rectangle, mask image.Image, maskp image.Point, advance fixed.Int26_6, ok bool) {
	cf.mu.Lock()
	defer cf.mu.Unlock()
	return cf.face.Glyph(dot, r)
}

func (cf *concurrentFace) GlyphBounds(r rune) (bounds fixed.Rectangle26_6, advance fixed.Int26_6, ok bool) {
	cf.mu.Lock()
	defer cf.mu.Unlock()
	return cf.face.GlyphBounds(r)
}

func (cf *concurrentFace) GlyphAdvance(r rune) (advance fixed.Int26_6, ok bool) {
	cf.mu.Lock()
	defer cf.mu.Unlock()
	return cf.face.GlyphAdvance(r)
}

func (cf *concurrentFace) Kern(r0, r1 rune) fixed.Int26_6 {
	cf.mu.Lock()
	defer cf.mu.Unlock()
	return cf.face.Kern(r0, r1)
}

func (cf *concurrentFace) Metrics() font.Metrics {
	cf.mu.Lock()
	defer cf.mu.Unlock()
	return cf.face.Metrics()
}

func TTFToFace(ttf []byte, size float64) (font.Face, error) {
	font, err := truetype.Parse(ttf)
	if err != nil {
		return nil, err
	}
	return &concurrentFace{face: truetype.NewFace(font, &truetype.Options{
		Size: size,
	})}, nil
}

func MakeTextImage(text string, face font.Face, clr color.Color) image.Image {
	drawer := &font.Drawer{
		Src:  &image.Uniform{clr},
		Face: face,
		Dot:  fixed.P(0, 0),
	}
	b26_6, _ := drawer.BoundString(text)
	bounds := image.Rect(
		b26_6.Min.X.Floor(),
		b26_6.Min.Y.Floor(),
		b26_6.Max.X.Ceil(),
		b26_6.Max.Y.Ceil(),
	)
	drawer.Dst = image.NewRGBA(bounds)
	drawer.DrawString(text)
	return drawer.Dst
}

func DrawCentered(dst draw.Image, r image.Rectangle, src image.Image, op draw.Op) {
	if src == nil {
		return
	}
	bounds := src.Bounds()
	center := bounds.Min.Add(bounds.Max).Div(2)
	target := r.Min.Add(r.Max).Div(2)
	delta := target.Sub(center)
	draw.Draw(dst, bounds.Add(delta).Intersect(r), src, bounds.Min, op)
}

func DrawLeftCentered(dst draw.Image, r image.Rectangle, src image.Image, op draw.Op) {
	if src == nil {
		return
	}
	bounds := src.Bounds()
	leftCenter := image.Pt(bounds.Min.X, (bounds.Min.Y+bounds.Max.Y)/2)
	target := image.Pt(r.Min.X, (r.Min.Y+r.Max.Y)/2)
	delta := target.Sub(leftCenter)
	draw.Draw(dst, bounds.Add(delta).Intersect(r), src, bounds.Min, op)
}
