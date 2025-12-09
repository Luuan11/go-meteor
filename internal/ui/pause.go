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

	pauseStatsStartY = 210
	pauseStatsSpacing = 25
	pauseMenuStartY = 340
	pauseMenuSpacing = 80
)

var (
	colorPauseWhite = color.RGBA{255, 255, 255, 200}
	colorPauseBlue  = color.RGBA{100, 200, 255, 200}
	colorPauseRed   = color.RGBA{255, 150, 50, 200}
	colorPausePurple = color.RGBA{200, 100, 255, 200}
	colorPauseGreen = color.RGBA{100, 255, 100, 200}
	colorPauseGold  = color.RGBA{255, 200, 0, 255}
	colorPauseHighlight = color.RGBA{100, 100, 100, 100}
)

type PauseMenu struct {
	selectedOption    int
	options           []string
	settingsButton    *IconButton
	shopButton        *IconButton
	score             int
	wave              int
	meteorsDestroyed  int
	powerUpsCollected int
	survivalTime      time.Duration
}

func NewPauseMenu() *PauseMenu {
	return &PauseMenu{
		options: []string{"Continue", "Restart", "Quit"},
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

func (p *PauseMenu) Draw(screen *ebiten.Image) {
	p.drawOverlay(screen)
	p.drawTitle(screen)
	p.drawStatistics(screen)
	p.drawMenuOptions(screen)
	p.drawButtons(screen)
}

func (p *PauseMenu) drawOverlay(screen *ebiten.Image) {
	overlay := ebiten.NewImage(config.ScreenWidth, config.ScreenHeight)
	overlay.Fill(color.RGBA{0, 0, 0, 180})
	screen.DrawImage(overlay, nil)
}

func (p *PauseMenu) drawTitle(screen *ebiten.Image) {
	titleText := "PAUSED"
	titleBounds := text.BoundString(assets.FontUi, titleText)
	titleX := (config.ScreenWidth - titleBounds.Dx()) / 2
	text.Draw(screen, titleText, assets.FontUi, titleX, 150, colorPauseWhite)
}

func (p *PauseMenu) drawStatistics(screen *ebiten.Image) {
	statsX := (config.ScreenWidth - 300) / 2
	y := pauseStatsStartY

	stats := []struct {
		text  string
		color color.Color
	}{
		{fmt.Sprintf("Score: %d", p.score), colorPauseWhite},
		{fmt.Sprintf("Wave: %d", p.wave), colorPauseBlue},
		{fmt.Sprintf("Meteors: %d", p.meteorsDestroyed), colorPauseRed},
		{fmt.Sprintf("Power-ups: %d", p.powerUpsCollected), colorPausePurple},
		{p.formatSurvivalTime(), colorPauseGreen},
	}

	for _, stat := range stats {
		text.Draw(screen, stat.text, assets.FontSmall, statsX, y, stat.color)
		y += pauseStatsSpacing
	}
}

func (p *PauseMenu) formatSurvivalTime() string {
	minutes := int(p.survivalTime.Minutes())
	seconds := int(p.survivalTime.Seconds()) % 60
	return fmt.Sprintf("Time: %02d:%02d", minutes, seconds)
}

func (p *PauseMenu) drawMenuOptions(screen *ebiten.Image) {
	for i, option := range p.options {
		optionColor := colorPauseWhite
		if i == p.selectedOption {
			optionColor = colorPauseGold
			p.drawOptionHighlight(screen, i)
		}

		optionBounds := text.BoundString(assets.FontUi, option)
		optionX := (config.ScreenWidth - optionBounds.Dx()) / 2
		optionY := pauseMenuStartY + i*pauseMenuSpacing
		text.Draw(screen, option, assets.FontUi, optionX, optionY+40, optionColor)
	}
}

func (p *PauseMenu) drawOptionHighlight(screen *ebiten.Image, index int) {
	x := float32((config.ScreenWidth - 300) / 2)
	y := float32(pauseMenuStartY + index*pauseMenuSpacing)
	vector.DrawFilledRect(screen, x, y, 300, 60, colorPauseHighlight, false)
}

func (p *PauseMenu) drawButtons(screen *ebiten.Image) {
	p.drawIconButton(screen, p.settingsButton)
	p.drawIconButton(screen, p.shopButton)
}

func (p *PauseMenu) drawIconButton(screen *ebiten.Image, btn *IconButton) {
	x, y := ebiten.CursorPosition()
	op := &ebiten.DrawImageOptions{}
	if p.isButtonHovered(btn, float64(x), float64(y)) {
		op.ColorScale.ScaleWithColor(colorPauseGold)
	}
	op.GeoM.Translate(btn.x, btn.y)
	screen.DrawImage(btn.icon, op)
}

func (p *PauseMenu) isButtonHovered(btn *IconButton, x, y float64) bool {
	return x >= btn.x && x <= btn.x+btn.size && y >= btn.y && y <= btn.y+btn.size
}

func (p *PauseMenu) Update() int {
	if action := p.handleKeyboardInput(); action != PauseActionNone {
		return action
	}
	return p.handleMouseAndTouch()
}

func (p *PauseMenu) handleKeyboardInput() int {
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		p.selectedOption = (p.selectedOption + 1) % len(p.options)
	}
	
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		p.selectedOption--
		if p.selectedOption < 0 {
			p.selectedOption = len(p.options) - 1
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		return p.selectedOption
	}

	return PauseActionNone
}

func (p *PauseMenu) handleMouseAndTouch() int {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return p.handleClick(ebiten.CursorPosition())
	}

	if touchIDs := inpututil.AppendJustPressedTouchIDs(nil); len(touchIDs) > 0 {
		return p.handleClick(ebiten.TouchPosition(touchIDs[0]))
	}

	return PauseActionNone
}

func (p *PauseMenu) handleClick(x, y int) int {
	if p.isButtonHovered(p.settingsButton, float64(x), float64(y)) {
		return PauseActionSettings
	}

	if p.isButtonHovered(p.shopButton, float64(x), float64(y)) {
		return PauseActionShop
	}

	return PauseActionNone
}
