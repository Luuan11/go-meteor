package ui

import (
	"fmt"
	"go-meteor/internal/config"
	assets "go-meteor/src/pkg"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type BossBar struct {
	visible bool
}

func NewBossBar() *BossBar {
	return &BossBar{
		visible: false,
	}
}

func (bb *BossBar) Show() {
	bb.visible = true
}

func (bb *BossBar) Hide() {
	bb.visible = false
}

func (bb *BossBar) Draw(screen *ebiten.Image, currentHealth, maxHealth int) {
	if !bb.visible {
		return
	}

	barWidth := float32(600)
	barHeight := float32(30)
	x := float32(config.ScreenWidth-int(barWidth)) / 2
	y := float32(60)

	bgColor := color.RGBA{50, 50, 50, 200}
	vector.DrawFilledRect(screen, x-2, y-2, barWidth+4, barHeight+4, bgColor, false)

	emptyColor := color.RGBA{100, 30, 30, 255}
	vector.DrawFilledRect(screen, x, y, barWidth, barHeight, emptyColor, false)

	healthPercent := float32(currentHealth) / float32(maxHealth)
	if healthPercent < 0 {
		healthPercent = 0
	}
	currentWidth := barWidth * healthPercent

	var healthColor color.RGBA
	if healthPercent > 0.6 {
		healthColor = color.RGBA{50, 255, 50, 255}
	} else if healthPercent > 0.3 {
		healthColor = color.RGBA{255, 200, 50, 255}
	} else {
		healthColor = color.RGBA{255, 50, 50, 255}
	}

	vector.DrawFilledRect(screen, x, y, currentWidth, barHeight, healthColor, false)

	borderColor := color.RGBA{200, 200, 200, 255}
	vector.StrokeRect(screen, x, y, barWidth, barHeight, 2, borderColor, false)

	bossText := "⚠️ ALIEN BOSS"
	bossX := int(x)
	text.Draw(screen, bossText, assets.FontSmall, bossX, int(y)-5, color.RGBA{255, 50, 50, 255})

	healthText := fmt.Sprintf("%d/%d HP", currentHealth, maxHealth)
	healthBounds := text.BoundString(assets.FontSmall, healthText)
	healthX := int(x+barWidth) - healthBounds.Dx()
	text.Draw(screen, healthText, assets.FontSmall, healthX, int(y)-5, color.White)
}
