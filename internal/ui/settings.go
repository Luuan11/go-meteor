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

const (
	settingsOptionMasterVolume = 0
	settingsOptionSFXVolume    = 1
	settingsOptionToggleSFX    = 2
	settingsOptionBack         = 3
	settingsTotalOptions       = 4

	settingsStartY  = 150
	settingsSpacing = 55
	settingsButtonSize = 35
)

var (
	settingsOverlay     *ebiten.Image
	colorSettingsGold   = color.RGBA{255, 215, 0, 255}
	colorSettingsGray   = color.RGBA{180, 180, 180, 255}
	colorSettingsWhite  = color.RGBA{255, 255, 255, 255}
	colorBtnMinus       = color.RGBA{150, 50, 50, 200}
	colorBtnMinusHover  = color.RGBA{200, 80, 80, 255}
	colorBtnPlus        = color.RGBA{50, 150, 50, 200}
	colorBtnPlusHover   = color.RGBA{80, 200, 80, 255}
)

func init() {
	settingsOverlay = ebiten.NewImage(config.ScreenWidth, config.ScreenHeight)
	settingsOverlay.Fill(color.RGBA{0, 0, 0, 200})
}

type Settings struct {
	selectedOption int
	cooldown       int
	closed         bool
}

func NewSettings() *Settings {
	return &Settings{}
}

func (s *Settings) Draw(screen *ebiten.Image) {
	screen.DrawImage(settingsOverlay, nil)
	s.drawTitle(screen)
	s.drawOptions(screen)
}

func (s *Settings) drawTitle(screen *ebiten.Image) {
	titleText := "AUDIO SETTINGS"
	titleBounds := text.BoundString(assets.FontUi, titleText)
	titleX := (config.ScreenWidth - titleBounds.Dx()) / 2
	text.Draw(screen, titleText, assets.FontUi, titleX, 80, colorSettingsWhite)
}

func (s *Settings) drawOptions(screen *ebiten.Image) {
	s.drawVolumeOption(screen, settingsOptionMasterVolume, "Master Volume", assets.GetMasterVolume(), settingsStartY)
	s.drawVolumeOption(screen, settingsOptionSFXVolume, "SFX Volume", assets.GetSFXVolume(), settingsStartY+settingsSpacing)
	s.drawToggleOption(screen, settingsOptionToggleSFX, settingsStartY+settingsSpacing*2)
	s.drawBackOption(screen, settingsOptionBack, settingsStartY+settingsSpacing*3+20)
}

func (s *Settings) drawVolumeOption(screen *ebiten.Image, index int, label string, volume float64, y int) {
	optionColor := s.getOptionColor(index)
	labelText := fmt.Sprintf("%s: %.0f%%", label, volume*100)
	labelBounds := text.BoundString(assets.FontUi, labelText)
	labelX := (config.ScreenWidth - labelBounds.Dx()) / 2
	text.Draw(screen, labelText, assets.FontUi, labelX, y, optionColor)

	minusX := labelX - 40
	minusY := y - 30
	s.drawVolumeButton(screen, minusX, minusY, "-", colorBtnMinus, colorBtnMinusHover)

	plusX := labelX + labelBounds.Dx() + 5
	plusY := y - 30
	s.drawVolumeButton(screen, plusX, plusY, "+", colorBtnPlus, colorBtnPlusHover)
}

func (s *Settings) drawToggleOption(screen *ebiten.Image, index int, y int) {
	sfxStatus := "ON"
	if !assets.IsSFXEnabled() {
		sfxStatus = "OFF"
	}
	s.drawSimpleOption(screen, index, fmt.Sprintf("Sound Effects: %s", sfxStatus), y)
}

func (s *Settings) drawBackOption(screen *ebiten.Image, index int, y int) {
	s.drawSimpleOption(screen, index, "BACK", y)
}

func (s *Settings) drawSimpleOption(screen *ebiten.Image, index int, label string, y int) {
	optionColor := s.getOptionColor(index)
	labelBounds := text.BoundString(assets.FontUi, label)
	labelX := (config.ScreenWidth - labelBounds.Dx()) / 2
	text.Draw(screen, label, assets.FontUi, labelX, y, optionColor)
}

func (s *Settings) getOptionColor(index int) color.Color {
	if s.selectedOption == index {
		return colorSettingsGold
	}
	return colorSettingsGray
}

func (s *Settings) drawVolumeButton(screen *ebiten.Image, x, y int, label string, btnColor, hoverColor color.RGBA) {
	mouseX, mouseY := ebiten.CursorPosition()
	finalColor := btnColor
	if mouseX >= x && mouseX <= x+settingsButtonSize && mouseY >= y && mouseY <= y+settingsButtonSize {
		finalColor = hoverColor
	}

	btn := ebiten.NewImage(settingsButtonSize, settingsButtonSize)
	btn.Fill(finalColor)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(btn, op)

	textX := x + 5
	if label == "+" {
		textX = x + 3
	}
	text.Draw(screen, label, assets.FontUi, textX, y+33, colorSettingsWhite)
}

func (s *Settings) Update() {
	if s.cooldown > 0 {
		s.cooldown--
		return
	}

	s.handleKeyboardInput()
	s.handleMouseAndTouch()
}

func (s *Settings) handleKeyboardInput() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		s.closed = true
		s.cooldown = 10
		return
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		s.moveSelectionUp()
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		s.moveSelectionDown()
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

func (s *Settings) moveSelectionUp() {
	s.selectedOption--
	if s.selectedOption < 0 {
		s.selectedOption = settingsTotalOptions - 1
	}
	s.cooldown = 8
}

func (s *Settings) moveSelectionDown() {
	s.selectedOption++
	if s.selectedOption >= settingsTotalOptions {
		s.selectedOption = 0
	}
	s.cooldown = 8
}

func (s *Settings) handleMouseAndTouch() {
	handleClick := func(x, y int) {
		if s.handleVolumeButtonClick(x, y) {
			return
		}
		s.handleOptionClick(x, y)
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		handleClick(ebiten.CursorPosition())
	}

	if touchIDs := inpututil.AppendJustPressedTouchIDs(nil); len(touchIDs) > 0 {
		handleClick(ebiten.TouchPosition(touchIDs[0]))
	}
}

func (s *Settings) handleVolumeButtonClick(x, y int) bool {
	for i := 0; i < 2; i++ {
		yPos := settingsStartY + settingsSpacing*i

		labelText := s.getVolumeLabelText(i)
		labelBounds := text.BoundString(assets.FontUi, labelText)
		labelX := (config.ScreenWidth - labelBounds.Dx()) / 2

		minusX := labelX - 40
		minusY := yPos - 25
		if x >= minusX && x <= minusX+settingsButtonSize && y >= minusY && y <= minusY+settingsButtonSize {
			s.selectedOption = i
			s.adjustOption(-0.1)
			s.cooldown = 5
			return true
		}

		plusX := labelX + labelBounds.Dx() + 5
		plusY := yPos - 25
		if x >= plusX && x <= plusX+settingsButtonSize && y >= plusY && y <= plusY+settingsButtonSize {
			s.selectedOption = i
			s.adjustOption(0.1)
			s.cooldown = 5
			return true
		}
	}
	return false
}

func (s *Settings) getVolumeLabelText(index int) string {
	if index == 0 {
		return fmt.Sprintf("Master Volume: %.0f%%", assets.GetMasterVolume()*100)
	}
	return fmt.Sprintf("SFX Volume: %.0f%%", assets.GetSFXVolume()*100)
}

func (s *Settings) handleOptionClick(x, y int) {
	for i := 0; i < settingsTotalOptions; i++ {
		yPos := settingsStartY + settingsSpacing*i
		if i == 3 {
			yPos += 20
		}

		if y >= yPos-20 && y <= yPos+20 && x >= config.ScreenWidth/4 && x <= config.ScreenWidth*3/4 {
			s.selectedOption = i
			s.activateOption()
			s.cooldown = 10
			break
		}
	}
}

func (s *Settings) adjustOption(delta float64) {
	switch s.selectedOption {
	case settingsOptionMasterVolume:
		assets.SetMasterVolume(assets.GetMasterVolume() + delta)
	case settingsOptionSFXVolume:
		assets.SetSFXVolume(assets.GetSFXVolume() + delta)
	}
}

func (s *Settings) activateOption() {
	switch s.selectedOption {
	case settingsOptionToggleSFX:
		assets.ToggleSFX()
	case settingsOptionBack:
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
