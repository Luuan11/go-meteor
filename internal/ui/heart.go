package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

var (
	heartFilledColor   = color.RGBA{255, 50, 50, 255}
	heartUnfilledColor = color.RGBA{100, 100, 100, 100}
)

var heartFilledSprite *ebiten.Image
var heartUnfilledSprite *ebiten.Image
var heartSpritesInitialized = false

func initHeartSprites() {
	if heartSpritesInitialized {
		return
	}

	heartPixels := []struct{ dx, dy int }{
		{2, 0}, {3, 0}, {5, 0}, {6, 0},
		{1, 1}, {2, 1}, {3, 1}, {4, 1}, {5, 1}, {6, 1}, {7, 1},
		{0, 2}, {1, 2}, {2, 2}, {3, 2}, {4, 2}, {5, 2}, {6, 2}, {7, 2}, {8, 2},
		{0, 3}, {1, 3}, {2, 3}, {3, 3}, {4, 3}, {5, 3}, {6, 3}, {7, 3}, {8, 3},
		{0, 4}, {1, 4}, {2, 4}, {3, 4}, {4, 4}, {5, 4}, {6, 4}, {7, 4}, {8, 4},
		{1, 5}, {2, 5}, {3, 5}, {4, 5}, {5, 5}, {6, 5}, {7, 5},
		{2, 6}, {3, 6}, {4, 6}, {5, 6}, {6, 6},
		{3, 7}, {4, 7}, {5, 7},
		{4, 8},
	}

	scale := 2
	width := 9 * scale
	height := 9 * scale

	heartFilledSprite = ebiten.NewImage(width, height)
	for _, p := range heartPixels {
		for sx := 0; sx < scale; sx++ {
			for sy := 0; sy < scale; sy++ {
				heartFilledSprite.Set(p.dx*scale+sx, p.dy*scale+sy, heartFilledColor)
			}
		}
	}

	heartUnfilledSprite = ebiten.NewImage(width, height)
	for _, p := range heartPixels {
		for sx := 0; sx < scale; sx++ {
			for sy := 0; sy < scale; sy++ {
				heartUnfilledSprite.Set(p.dx*scale+sx, p.dy*scale+sy, heartUnfilledColor)
			}
		}
	}

	heartSpritesInitialized = true
}

func DrawHeart(screen *ebiten.Image, x, y int, filled bool) {
	initHeartSprites()

	sprite := heartUnfilledSprite
	if filled {
		sprite = heartFilledSprite
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(sprite, op)
}
