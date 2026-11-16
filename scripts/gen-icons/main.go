package main

import (
	"image"
	"image/color"
	"image/gif"
	"math"
	"os"
	"path/filepath"
)

func main() {
	outDir := filepath.Join("qpkg", "icons")
	if len(os.Args) > 1 {
		outDir = os.Args[1]
	}
	if err := os.MkdirAll(outDir, 0755); err != nil {
		panic(err)
	}

	writeGIF(filepath.Join(outDir, "BeszelAgent.gif"), renderIcon(64, false))
	writeGIF(filepath.Join(outDir, "BeszelAgent_80.gif"), renderIcon(80, false))
	writeGIF(filepath.Join(outDir, "BeszelAgent_gray.gif"), renderIcon(64, true))
}

func writeGIF(path string, img image.Image) {
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err := gif.Encode(f, img, &gif.Options{NumColors: 256}); err != nil {
		panic(err)
	}
}

func renderIcon(size int, gray bool) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	s := float64(size)
	r := s * 0.22
	margin := s * 0.05

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			fx, fy := float64(x)+0.5, float64(y)+0.5
			if !inRoundedRect(fx, fy, margin, margin, s-2*margin, s-2*margin, r) {
				img.SetRGBA(x, y, color.RGBA{0, 0, 0, 0})
				continue
			}

			// Gradient background: sky-500 -> blue-700
			t := (fx + fy) / (2.0 * s)
			c1 := color.RGBA{14, 165, 233, 255}
			c2 := color.RGBA{3, 105, 161, 255}
			bg := lerp(c1, c2, t)

			// Draw pulse/chart line and base bar normalized to [0, 1]
			nx := fx / s
			ny := fy / s
			onPulse := distToSegments(nx, ny, [][2]float64{
				{0.22, 0.56},
				{0.39, 0.39},
				{0.52, 0.52},
				{0.77, 0.27},
			}) < 0.048
			onBar := ny >= 0.68 && ny <= 0.75 && nx >= 0.22 && nx <= 0.78

			px := bg
			if onPulse {
				px = color.RGBA{255, 255, 255, 255}
			} else if onBar {
				px = color.RGBA{186, 230, 253, 255}
			}
			if gray {
				lum := uint8(0.299*float64(px.R) + 0.587*float64(px.G) + 0.114*float64(px.B))
				px = color.RGBA{lum, lum, lum, px.A}
			}
			img.SetRGBA(x, y, px)
		}
	}
	return img
}

func lerp(a, b color.RGBA, t float64) color.RGBA {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	return color.RGBA{
		R: uint8(float64(a.R) + (float64(b.R)-float64(a.R))*t),
		G: uint8(float64(a.G) + (float64(b.G)-float64(a.G))*t),
		B: uint8(float64(a.B) + (float64(b.B)-float64(a.B))*t),
		A: 255,
	}
}

func inRoundedRect(px, py, x, y, w, h, r float64) bool {
	if px < x || px > x+w || py < y || py > y+h {
		return false
	}
	cx := math.Max(x+r, math.Min(px, x+w-r))
	cy := math.Max(y+r, math.Min(py, y+h-r))
	dx, dy := px-cx, py-cy
	return dx*dx+dy*dy <= r*r
}

func distToSegments(px, py float64, pts [][2]float64) float64 {
	best := math.MaxFloat64
	for i := 0; i < len(pts)-1; i++ {
		d := distToSeg(px, py, pts[i][0], pts[i][1], pts[i+1][0], pts[i+1][1])
		if d < best {
			best = d
		}
	}
	return best
}

func distToSeg(px, py, x1, y1, x2, y2 float64) float64 {
	vx, vy := x2-x1, y2-y1
	wx, wy := px-x1, py-y1
	c1 := wx*vx + wy*vy
	if c1 <= 0 {
		return math.Hypot(px-x1, py-y1)
	}
	c2 := vx*vx + vy*vy
	if c2 <= c1 {
		return math.Hypot(px-x2, py-y2)
	}
	b := c1 / c2
	bx, by := x1+b*vx, y1+b*vy
	return math.Hypot(px-bx, py-by)
}
