package ui

import (
	"fmt"
	"go-meteor/internal/config"
	assets "go-meteor/src/pkg"
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
)

type Statistics struct {
	meteorsDestroyed  int
	powerUpsCollected int
	survivalTime      time.Duration
	wave              int
	score             int
	settingsButton    *IconButton
	shopButton        *IconButton
	openSettings      bool
	openShop          bool
}

func NewStatistics(meteors, powerUps, wave, score int, survival time.Duration) *Statistics {
	return &Statistics{
		meteorsDestroyed:  meteors,
		powerUpsCollected: powerUps,
		survivalTime:      survival,
		wave:              wave,
		score:             score,
		openSettings:      false,
		openShop:          false,
		settingsButton: &IconButton{
			x:    config.ScreenWidth - 50,
			y:    10,
			size: 35,
			icon: CreateGearIcon(35),
		},
		shopButton: &IconButton{
			x:    10,
			y:    10,
			size: 35,
			icon: assets.CoinSprite,
		},
	}
}

func (s *Statistics) Draw(screen *ebiten.Image, startY int) {
	titleText := "GAME STATISTICS"
	titleBounds := text.BoundString(assets.FontSmall, titleText)
	titleX := (config.ScreenWidth - titleBounds.Dx()) / 2
	text.Draw(screen, titleText, assets.FontSmall, titleX, startY, color.RGBA{255, 215, 0, 255})

	statsY := startY + 30
	lineSpacing := 24

	// Score
	scoreText := fmt.Sprintf("Final Score: %d", s.score)
	scoreBounds := text.BoundString(assets.FontSmall, scoreText)
	scoreX := (config.ScreenWidth - scoreBounds.Dx()) / 2
	text.Draw(screen, scoreText, assets.FontSmall, scoreX, statsY, color.White)
	statsY += lineSpacing

	// Wave
	waveText := fmt.Sprintf("Waves Completed: %d", s.wave-1)
	waveBounds := text.BoundString(assets.FontSmall, waveText)
	waveX := (config.ScreenWidth - waveBounds.Dx()) / 2
	text.Draw(screen, waveText, assets.FontSmall, waveX, statsY, color.RGBA{150, 200, 255, 255})
	statsY += lineSpacing

	// Meteors
	meteorText := fmt.Sprintf("Meteors Destroyed: %d", s.meteorsDestroyed)
	meteorBounds := text.BoundString(assets.FontSmall, meteorText)
	meteorX := (config.ScreenWidth - meteorBounds.Dx()) / 2
	text.Draw(screen, meteorText, assets.FontSmall, meteorX, statsY, color.RGBA{255, 150, 100, 255})
	statsY += lineSpacing

	// Power-ups
	powerUpText := fmt.Sprintf("Power-Ups Collected: %d", s.powerUpsCollected)
	powerUpBounds := text.BoundString(assets.FontSmall, powerUpText)
	powerUpX := (config.ScreenWidth - powerUpBounds.Dx()) / 2
	text.Draw(screen, powerUpText, assets.FontSmall, powerUpX, statsY, color.RGBA{255, 200, 255, 255})
	statsY += lineSpacing

	// Survival time
	minutes := int(s.survivalTime.Minutes())
	seconds := int(s.survivalTime.Seconds()) % 60
	timeText := fmt.Sprintf("Survival Time: %02d:%02d", minutes, seconds)
	timeBounds := text.BoundString(assets.FontSmall, timeText)
	timeX := (config.ScreenWidth - timeBounds.Dx()) / 2
	text.Draw(screen, timeText, assets.FontSmall, timeX, statsY, color.RGBA{100, 255, 100, 255})

	s.drawSettingsButton(screen)
	s.drawShopButton(screen)
}

func (s *Statistics) drawSettingsButton(screen *ebiten.Image) {
	btn := s.settingsButton
	mouseX, mouseY := ebiten.CursorPosition()
	isHovered := float64(mouseX) >= btn.x && float64(mouseX) <= btn.x+btn.size &&
		float64(mouseY) >= btn.y && float64(mouseY) <= btn.y+btn.size

	op := &ebiten.DrawImageOptions{}
	if isHovered {
		op.ColorScale.ScaleWithColor(color.RGBA{255, 215, 0, 255})
	}
	op.GeoM.Translate(btn.x, btn.y)
	screen.DrawImage(btn.icon, op)
}

func (s *Statistics) drawShopButton(screen *ebiten.Image) {
	btn := s.shopButton
	mouseX, mouseY := ebiten.CursorPosition()
	isHovered := float64(mouseX) >= btn.x && float64(mouseX) <= btn.x+70 &&
		float64(mouseY) >= btn.y && float64(mouseY) <= btn.y+btn.size

	op := &ebiten.DrawImageOptions{}
	if isHovered {
		op.ColorScale.ScaleWithColor(color.RGBA{255, 215, 0, 255})
	}
	op.GeoM.Translate(btn.x, btn.y)
	screen.DrawImage(btn.icon, op)

	var shopTextColor color.Color = color.White
	if isHovered {
		shopTextColor = color.RGBA{255, 215, 0, 255}
	}
	text.Draw(screen, "Shop", assets.FontSmall, int(btn.x)+40, int(btn.y)+22, shopTextColor)
}

func (s *Statistics) CheckSettingsClick() bool {
	btn := s.settingsButton
	mouseX, mouseY := ebiten.CursorPosition()
	return float64(mouseX) >= btn.x && float64(mouseX) <= btn.x+btn.size &&
		float64(mouseY) >= btn.y && float64(mouseY) <= btn.y+btn.size
}

func (s *Statistics) CheckShopClick() bool {
	btn := s.shopButton
	mouseX, mouseY := ebiten.CursorPosition()
	return float64(mouseX) >= btn.x && float64(mouseX) <= btn.x+70 &&
		float64(mouseY) >= btn.y && float64(mouseY) <= btn.y+btn.size
}
