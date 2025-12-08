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

var settingsOverlay *ebiten.Image

func init() {
	settingsOverlay = ebiten.NewImage(config.ScreenWidth, config.ScreenHeight)
	settingsOverlay.Fill(color.RGBA{0, 0, 0, 200})
}

type Settings struct {
	selectedOption int
	cooldown       int
	closed         bool
}

const (
	optionMasterVolume = 0
	optionSFXVolume    = 1
	optionToggleSFX    = 2
	optionBack         = 3
	totalOptions       = 4
)

func NewSettings() *Settings {
	return &Settings{
		selectedOption: 0,
		cooldown:       0,
		closed:         false,
	}
}

func (s *Settings) Draw(screen *ebiten.Image) {
	screen.DrawImage(settingsOverlay, nil)

	titleText := "AUDIO SETTINGS"
	titleBounds := text.BoundString(assets.FontUi, titleText)
	titleX := (config.ScreenWidth - titleBounds.Dx()) / 2
	text.Draw(screen, titleText, assets.FontUi, titleX, 80, color.White)

	optionY := 150
	spacing := 55

	s.drawVolumeOption(screen, 0, "Master Volume", assets.GetMasterVolume(), optionY+spacing*0)
	s.drawVolumeOption(screen, 1, "SFX Volume", assets.GetSFXVolume(), optionY+spacing*1)

	sfxStatus := "ON"
	if !assets.IsSFXEnabled() {
		sfxStatus = "OFF"
	}
	s.drawOption(screen, 2, fmt.Sprintf("Sound Effects: %s", sfxStatus), optionY+spacing*2)
	s.drawOption(screen, 3, "BACK", optionY+spacing*3+20)
}

func (s *Settings) drawVolumeButton(screen *ebiten.Image, x, y int, label string, btnColor, hoverColor color.RGBA) {
	mouseX, mouseY := ebiten.CursorPosition()
	finalColor := btnColor
	if mouseX >= x && mouseX <= x+35 && mouseY >= y && mouseY <= y+35 {
		finalColor = hoverColor
	}
	btn := ebiten.NewImage(35, 35)
	btn.Fill(finalColor)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(btn, op)
	textX := x + 5
	if label == "+" {
		textX = x + 3
	}
	text.Draw(screen, label, assets.FontUi, textX, y+33, color.White)
}

func (s *Settings) drawVolumeOption(screen *ebiten.Image, index int, label string, volume float64, y int) {
	optionColor := color.RGBA{180, 180, 180, 255}
	if s.selectedOption == index {
		optionColor = color.RGBA{255, 215, 0, 255}
	}

	labelText := fmt.Sprintf("%s: %.0f%%", label, volume*100)
	labelBounds := text.BoundString(assets.FontUi, labelText)
	labelX := (config.ScreenWidth - labelBounds.Dx()) / 2
	text.Draw(screen, labelText, assets.FontUi, labelX, y, optionColor)

	minusX := labelX - 40
	minusY := y - 30
	s.drawVolumeButton(screen, minusX, minusY, "-",
		color.RGBA{150, 50, 50, 200},
		color.RGBA{200, 80, 80, 255})

	plusX := labelX + labelBounds.Dx() + 5
	plusY := y - 30
	s.drawVolumeButton(screen, plusX, plusY, "+",
		color.RGBA{50, 150, 50, 200},
		color.RGBA{80, 200, 80, 255})
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

	handleClick := func(clickX, clickY int) {
		optionY := 150
		spacing := 55

		for i := 0; i < 2; i++ {
			yPos := optionY + spacing*i

			volumeLabel := ""
			volume := 0.0
			if i == 0 {
				volumeLabel = "Master Volume"
				volume = assets.GetMasterVolume()
			} else {
				volumeLabel = "SFX Volume"
				volume = assets.GetSFXVolume()
			}

			labelText := fmt.Sprintf("%s: %.0f%%", volumeLabel, volume*100)
			labelBounds := text.BoundString(assets.FontUi, labelText)
			labelX := (config.ScreenWidth - labelBounds.Dx()) / 2

			// Check - button (updated position)
			minusX := labelX - 40
			minusY := yPos - 25
			if clickX >= minusX && clickX <= minusX+35 && clickY >= minusY && clickY <= minusY+35 {
				s.selectedOption = i
				s.adjustOption(-0.1)
				s.cooldown = 5
				return
			}

			// Check + button (updated position)
			plusX := labelX + labelBounds.Dx() + 5
			plusY := yPos - 25
			if clickX >= plusX && clickX <= plusX+35 && clickY >= plusY && clickY <= plusY+35 {
				s.selectedOption = i
				s.adjustOption(0.1)
				s.cooldown = 5
				return
			}
		}

		// Check toggle and back buttons
		for i := 0; i < totalOptions; i++ {
			yPos := optionY + spacing*i
			if i == 3 {
				yPos += 20
			}

			if clickY >= yPos-20 && clickY <= yPos+20 && clickX >= config.ScreenWidth/4 && clickX <= config.ScreenWidth*3/4 {
				s.selectedOption = i
				s.activateOption()
				s.cooldown = 10
				break
			}
		}
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mouseX, mouseY := ebiten.CursorPosition()
		handleClick(mouseX, mouseY)
	}

	touchIDs := inpututil.AppendJustPressedTouchIDs(nil)
	if len(touchIDs) > 0 {
		touchX, touchY := ebiten.TouchPosition(touchIDs[0])
		handleClick(touchX, touchY)
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
	}
}

func (s *Settings) activateOption() {
	switch s.selectedOption {
	case optionToggleSFX:
		assets.ToggleSFX()
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
