package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"go-meteor/internal/config"
)

func DrawPowerUpBar(screen *ebiten.Image, progress float32, barColor color.Color) {
	DrawPowerUpBarAt(screen, progress, barColor, 100)
}

func DrawPowerUpBarAt(screen *ebiten.Image, progress float32, barColor color.Color, barY float32) {
	barWidth := float32(200)
	barHeight := float32(20)
	barX := float32(config.ScreenWidth)/2 - barWidth/2

	vector.DrawFilledRect(screen, barX-2, barY-2, barWidth+4, barHeight+4, color.RGBA{50, 50, 50, 200}, false)
	vector.DrawFilledRect(screen, barX, barY, barWidth, barHeight, color.RGBA{30, 30, 30, 200}, false)

	fillWidth := barWidth * progress
	vector.DrawFilledRect(screen, barX, barY, fillWidth, barHeight, barColor, false)
}
