package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func DrawPauseIcon(screen *ebiten.Image, x, y int) {
	iconSize := 30
	barWidth := 8
	gap := 6

	bgColor := color.RGBA{0, 0, 0, 150}
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(iconSize), float32(iconSize), bgColor, false)

	barColor := color.RGBA{255, 255, 255, 200}

	leftBarX := x + (iconSize-barWidth*2-gap)/2
	vector.DrawFilledRect(screen, float32(leftBarX), float32(y+6), float32(barWidth), float32(iconSize-12), barColor, false)

	rightBarX := leftBarX + barWidth + gap
	vector.DrawFilledRect(screen, float32(rightBarX), float32(y+6), float32(barWidth), float32(iconSize-12), barColor, false)

	borderColor := color.RGBA{255, 255, 255, 100}
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(iconSize), 2, borderColor, false)
	vector.DrawFilledRect(screen, float32(x), float32(y+iconSize-2), float32(iconSize), 2, borderColor, false)
	vector.DrawFilledRect(screen, float32(x), float32(y), 2, float32(iconSize), borderColor, false)
	vector.DrawFilledRect(screen, float32(x+iconSize-2), float32(y), 2, float32(iconSize), borderColor, false)
}
