package ui

import (
	"fmt"
	"go-meteor/internal/config"
	assets "go-meteor/src/pkg"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
)

type Settings struct {
	selectedOption int
	cooldown       int
	closed         bool
}

const (
	optionMasterVolume = 0
	optionSFXVolume    = 1
	optionMusicVolume  = 2
	optionToggleSFX    = 3
	optionToggleMusic  = 4
	optionBack         = 5
	totalOptions       = 6
)

func NewSettings() *Settings {
	return &Settings{
		selectedOption: 0,
		cooldown:       0,
		closed:         false,
	}
}

func (s *Settings) Draw(screen *ebiten.Image) {
	overlay := ebiten.NewImage(config.ScreenWidth, config.ScreenHeight)
	overlay.Fill(color.RGBA{0, 0, 0, 200})
	screen.DrawImage(overlay, nil)

	titleText := "AUDIO SETTINGS"
	titleBounds := text.BoundString(assets.FontUi, titleText)
	titleX := (config.ScreenWidth - titleBounds.Dx()) / 2
	text.Draw(screen, titleText, assets.FontUi, titleX, 80, color.White)

	optionY := 150
	spacing := 55

	s.drawOption(screen, 0, fmt.Sprintf("Master Volume: %.0f%%", assets.GetMasterVolume()*100), optionY+spacing*0)
	s.drawOption(screen, 1, fmt.Sprintf("SFX Volume: %.0f%%", assets.GetSFXVolume()*100), optionY+spacing*1)
	s.drawOption(screen, 2, fmt.Sprintf("Music Volume: %.0f%%", assets.GetMusicVolume()*100), optionY+spacing*2)

	sfxStatus := "ON"
	if !assets.IsSFXEnabled() {
		sfxStatus = "OFF"
	}
	s.drawOption(screen, 3, fmt.Sprintf("Sound Effects: %s", sfxStatus), optionY+spacing*3)

	musicStatus := "ON"
	if !assets.IsMusicEnabled() {
		musicStatus = "OFF"
	}
	s.drawOption(screen, 4, fmt.Sprintf("Music: %s", musicStatus), optionY+spacing*4)

	s.drawOption(screen, 5, "BACK", optionY+spacing*5+20)
}

func (s *Settings) drawOption(screen *ebiten.Image, index int, label string, y int) {
	optionColor := color.RGBA{180, 180, 180, 255}
	if s.selectedOption == index {
		optionColor = color.RGBA{255, 215, 0, 255}
	}

	labelBounds := text.BoundString(assets.FontUi, label)
	labelX := (config.ScreenWidth - labelBounds.Dx()) / 2
	text.Draw(screen, label, assets.FontUi, labelX, y, optionColor)
}

func (s *Settings) Update() {
	if s.cooldown > 0 {
		s.cooldown--
		return
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		s.closed = true
		s.cooldown = 10
		return
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		s.selectedOption--
		if s.selectedOption < 0 {
			s.selectedOption = totalOptions - 1
		}
		s.cooldown = 8
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		s.selectedOption++
		if s.selectedOption >= totalOptions {
			s.selectedOption = 0
		}
		s.cooldown = 8
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) || inpututil.IsKeyJustPressed(ebiten.KeyA) {
		s.adjustOption(-0.1)
		s.cooldown = 5
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) || inpututil.IsKeyJustPressed(ebiten.KeyD) {
		s.adjustOption(0.1)
		s.cooldown = 5
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		s.activateOption()
		s.cooldown = 10
	}
}

func (s *Settings) adjustOption(delta float64) {
	switch s.selectedOption {
	case optionMasterVolume:
		newVol := assets.GetMasterVolume() + delta
		assets.SetMasterVolume(newVol)
	case optionSFXVolume:
		newVol := assets.GetSFXVolume() + delta
		assets.SetSFXVolume(newVol)
	case optionMusicVolume:
		newVol := assets.GetMusicVolume() + delta
		assets.SetMusicVolume(newVol)
	}
}

func (s *Settings) activateOption() {
	switch s.selectedOption {
	case optionToggleSFX:
		assets.ToggleSFX()
	case optionToggleMusic:
		assets.ToggleMusic()
	case optionBack:
		s.closed = true
	}
}

func (s *Settings) IsClosed() bool {
	return s.closed
}

func (s *Settings) Reset() {
	s.closed = false
	s.selectedOption = 0
	s.cooldown = 10
}
