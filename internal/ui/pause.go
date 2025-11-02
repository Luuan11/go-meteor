package ui

import (
	"go-meteor/internal/config"
	assets "go-meteor/src/pkg"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
)

type PauseMenu struct {
	selectedOption int
	options        []string
}

func NewPauseMenu() *PauseMenu {
	return &PauseMenu{
		selectedOption: 0,
		options:        []string{"Continue", "Restart", "Quit"},
	}
}

func (pm *PauseMenu) Draw(screen *ebiten.Image) {
	overlay := ebiten.NewImage(config.ScreenWidth, config.ScreenHeight)
	overlay.Fill(color.RGBA{0, 0, 0, 180})
	screen.DrawImage(overlay, nil)

	titleText := "PAUSED"
	titleWidth := font.MeasureString(assets.FontUi, titleText)
	titleX := (config.ScreenWidth - titleWidth.Ceil()) / 2
	text.Draw(screen, titleText, assets.FontUi, titleX, 200, color.White)

	for i, option := range pm.options {
		var optionColor color.Color = color.White
		if i == pm.selectedOption {
			optionColor = color.RGBA{255, 200, 0, 255}
			vector.DrawFilledRect(screen, float32((config.ScreenWidth-300)/2), float32(280+i*80), 300, 60, color.RGBA{100, 100, 100, 100}, false)
		}

		optionWidth := font.MeasureString(assets.FontUi, option)
		optionX := (config.ScreenWidth - optionWidth.Ceil()) / 2
		text.Draw(screen, option, assets.FontUi, optionX, 320+i*80, optionColor)
	}
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

	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		return pm.selectedOption
	}

	return -1
}
