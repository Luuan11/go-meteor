package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

func CreateGearIcon(size int) *ebiten.Image {
	img := ebiten.NewImage(size, size)
	center := float64(size) / 2
	outerRadius := center * 0.9
	innerRadius := center * 0.4
	teethCount := 8

	fillCircle(img, int(center), int(center), int(outerRadius), color.RGBA{200, 200, 200, 255})

	for i := 0; i < teethCount; i++ {
		angle := float64(i) * 2 * 3.14159 / float64(teethCount)
		toothWidth := 0.15

		x1 := center + outerRadius*0.8*cos(angle-toothWidth)
		y1 := center + outerRadius*0.8*sin(angle-toothWidth)
		x2 := center + outerRadius*1.1*cos(angle-toothWidth/2)
		y2 := center + outerRadius*1.1*sin(angle-toothWidth/2)
		x3 := center + outerRadius*1.1*cos(angle+toothWidth/2)
		y3 := center + outerRadius*1.1*sin(angle+toothWidth/2)
		x4 := center + outerRadius*0.8*cos(angle+toothWidth)
		y4 := center + outerRadius*0.8*sin(angle+toothWidth)

		fillPolygon(img, []float64{x1, x2, x3, x4}, []float64{y1, y2, y3, y4}, color.RGBA{200, 200, 200, 255})
	}

	fillCircle(img, int(center), int(center), int(innerRadius), color.RGBA{50, 50, 50, 255})

	return img
}

func fillCircle(img *ebiten.Image, centerX, centerY, radius int, col color.Color) {
	for y := -radius; y <= radius; y++ {
		for x := -radius; x <= radius; x++ {
			if x*x+y*y <= radius*radius {
				img.Set(centerX+x, centerY+y, col)
			}
		}
	}
}

func fillPolygon(img *ebiten.Image, xs, ys []float64, col color.Color) {
	if len(xs) != len(ys) || len(xs) < 3 {
		return
	}

	minX, maxX := xs[0], xs[0]
	minY, maxY := ys[0], ys[0]
	for i := 1; i < len(xs); i++ {
		if xs[i] < minX {
			minX = xs[i]
		}
		if xs[i] > maxX {
			maxX = xs[i]
		}
		if ys[i] < minY {
			minY = ys[i]
		}
		if ys[i] > maxY {
			maxY = ys[i]
		}
	}

	for y := int(minY); y <= int(maxY); y++ {
		for x := int(minX); x <= int(maxX); x++ {
			if pointInPolygon(float64(x), float64(y), xs, ys) {
				img.Set(x, y, col)
			}
		}
	}
}

func pointInPolygon(x, y float64, xs, ys []float64) bool {
	n := len(xs)
	inside := false

	j := n - 1
	for i := 0; i < n; i++ {
		xi, yi := xs[i], ys[i]
		xj, yj := xs[j], ys[j]

		if ((yi > y) != (yj > y)) && (x < (xj-xi)*(y-yi)/(yj-yi)+xi) {
			inside = !inside
		}
		j = i
	}

	return inside
}

func cos(angle float64) float64 {
	// Simple cosine approximation
	const pi = 3.14159265358979323846
	angle = angle - float64(int(angle/(2*pi)))*(2*pi)

	if angle < 0 {
		angle = -angle
	}

	if angle > pi {
		angle = 2*pi - angle
	}

	x := angle
	return 1 - x*x/2 + x*x*x*x/24 - x*x*x*x*x*x/720
}

func sin(angle float64) float64 {
	const pi = 3.14159265358979323846
	return cos(angle - pi/2)
}
