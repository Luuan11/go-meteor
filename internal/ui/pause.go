package ui

import (
	"fmt"
	"go-meteor/internal/config"
	assets "go-meteor/src/pkg"
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	PauseActionContinue = 0
	PauseActionRestart  = 1
	PauseActionQuit     = 2
	PauseActionSettings = 3
	PauseActionShop     = 4
	PauseActionNone     = -1
)

type PauseMenu struct {
	selectedOption int
	options        []string
	settingsButton *IconButton
	shopButton     *IconButton

	score             int
	wave              int
	meteorsDestroyed  int
	powerUpsCollected int
	survivalTime      time.Duration
}

func NewPauseMenu() *PauseMenu {
	return &PauseMenu{
		selectedOption: 0,
		options:        []string{"Continue", "Restart", "Quit"},
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

func (p *PauseMenu) SetStats(score, wave, meteorsDestroyed, powerUpsCollected int, survivalTime time.Duration) {
	p.score = score
	p.wave = wave
	p.meteorsDestroyed = meteorsDestroyed
	p.powerUpsCollected = powerUpsCollected
	p.survivalTime = survivalTime
}

func (pm *PauseMenu) Draw(screen *ebiten.Image) {
	overlay := ebiten.NewImage(config.ScreenWidth, config.ScreenHeight)
	overlay.Fill(color.RGBA{0, 0, 0, 180})
	screen.DrawImage(overlay, nil)

	titleText := "PAUSED"
	titleBounds := text.BoundString(assets.FontUi, titleText)
	titleX := (config.ScreenWidth - titleBounds.Dx()) / 2
	text.Draw(screen, titleText, assets.FontUi, titleX, 150, color.White)

	// Draw current statistics
	statsY := 210
	statsX := (config.ScreenWidth - 300) / 2

	scoreText := fmt.Sprintf("Score: %d", pm.score)
	text.Draw(screen, scoreText, assets.FontSmall, statsX, statsY, color.RGBA{255, 255, 255, 200})

	waveText := fmt.Sprintf("Wave: %d", pm.wave)
	text.Draw(screen, waveText, assets.FontSmall, statsX, statsY+25, color.RGBA{100, 200, 255, 200})

	meteorsText := fmt.Sprintf("Meteors: %d", pm.meteorsDestroyed)
	text.Draw(screen, meteorsText, assets.FontSmall, statsX, statsY+50, color.RGBA{255, 150, 50, 200})

	powerUpsText := fmt.Sprintf("Power-ups: %d", pm.powerUpsCollected)
	text.Draw(screen, powerUpsText, assets.FontSmall, statsX, statsY+75, color.RGBA{200, 100, 255, 200})

	minutes := int(pm.survivalTime.Minutes())
	seconds := int(pm.survivalTime.Seconds()) % 60
	timeText := fmt.Sprintf("Time: %02d:%02d", minutes, seconds)
	text.Draw(screen, timeText, assets.FontSmall, statsX, statsY+100, color.RGBA{100, 255, 100, 200})

	// Draw menu options
	for i, option := range pm.options {
		var optionColor color.Color = color.White
		if i == pm.selectedOption {
			optionColor = color.RGBA{255, 200, 0, 255}
			vector.DrawFilledRect(screen, float32((config.ScreenWidth-300)/2), float32(340+i*80), 300, 60, color.RGBA{100, 100, 100, 100}, false)
		}

		optionBounds := text.BoundString(assets.FontUi, option)
		optionX := (config.ScreenWidth - optionBounds.Dx()) / 2
		text.Draw(screen, option, assets.FontUi, optionX, 380+i*80, optionColor)
	}

	pm.drawSettingsButton(screen)
	pm.drawShopButton(screen)
}

func (pm *PauseMenu) drawSettingsButton(screen *ebiten.Image) {
	btn := pm.settingsButton
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

func (pm *PauseMenu) drawShopButton(screen *ebiten.Image) {
	btn := pm.shopButton
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

func (pm *PauseMenu) Update() int {
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		pm.selectedOption = (pm.selectedOption + 1) % len(pm.options)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		pm.selectedOption--
		if pm.selectedOption < 0 {
			pm.selectedOption = len(pm.options) - 1
		}
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mouseX, mouseY := ebiten.CursorPosition()
		
		// Check settings button
		btn := pm.settingsButton
		if float64(mouseX) >= btn.x && float64(mouseX) <= btn.x+btn.size &&
			float64(mouseY) >= btn.y && float64(mouseY) <= btn.y+btn.size {
			return PauseActionSettings
		}
		
		// Check shop button
		shopBtn := pm.shopButton
		if float64(mouseX) >= shopBtn.x && float64(mouseX) <= shopBtn.x+shopBtn.size &&
			float64(mouseY) >= shopBtn.y && float64(mouseY) <= shopBtn.y+shopBtn.size {
			return PauseActionShop
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		return pm.selectedOption
	}

	return PauseActionNone
}
