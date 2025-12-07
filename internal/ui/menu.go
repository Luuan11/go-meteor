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

func formatScore(score int) string {
	return fmt.Sprintf("%d", score)
}

type Menu struct {
	readyToPlay    bool
	openSettings   bool
	openShop       bool
	cooldown       int
	highScore      int
	lastScore      int
	settingsButton *IconButton
	shopButton     *IconButton
}

type IconButton struct {
	x, y, size float64
	icon       *ebiten.Image
}

func NewMenu() *Menu {
	return &Menu{
		readyToPlay:  false,
		openSettings: false,
		openShop:     false,
		cooldown:     0,
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

func (m *Menu) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(0.5, 0.5)
	op.GeoM.Translate(280, 50)
	screen.DrawImage(assets.GopherPlayer, op)

	titleText := "Space GO"
	titleBounds := text.BoundString(assets.ScoreFont, titleText)
	titleX := (config.ScreenWidth - titleBounds.Dx()) / 2
	text.Draw(screen, titleText, assets.ScoreFont, titleX, 320, color.White)

	highScoreText := "High Score: " + formatScore(m.highScore)
	highScoreBounds := text.BoundString(assets.FontSmall, highScoreText)
	highScoreX := (config.ScreenWidth - highScoreBounds.Dx()) / 2
	text.Draw(screen, highScoreText, assets.FontSmall, highScoreX, 510, color.RGBA{255, 215, 0, 255})

	if m.lastScore > 0 {
		lastScoreText := "Last Score: " + formatScore(m.lastScore)
		lastScoreBounds := text.BoundString(assets.FontSmall, lastScoreText)
		lastScoreX := (config.ScreenWidth - lastScoreBounds.Dx()) / 2
		text.Draw(screen, lastScoreText, assets.FontSmall, lastScoreX, 535, color.RGBA{143, 47, 233, 255})
	}

	instructionText := "Press ENTER to start"
	instructionBounds := text.BoundString(assets.FontUi, instructionText)
	instructionX := (config.ScreenWidth - instructionBounds.Dx()) / 2
	text.Draw(screen, instructionText, assets.FontUi, instructionX, 400, color.White)

	creditText := "Luuan11"
	creditBounds := text.BoundString(assets.FontSmall, creditText)
	creditX := (config.ScreenWidth - creditBounds.Dx()) / 2
	text.Draw(screen, creditText, assets.FontSmall, creditX, 565, color.RGBA{150, 100, 255, 255})

	m.drawButton(screen, m.settingsButton)
	m.drawShopButton(screen)
}

func (m *Menu) drawButton(screen *ebiten.Image, btn *IconButton) {
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

func (m *Menu) drawShopButton(screen *ebiten.Image) {
	btn := m.shopButton
	mouseX, mouseY := ebiten.CursorPosition()

	// Check if mouse is hovering over button area (icon + text)
	textWidth := 70 // Approximate width for icon + "Shop" text
	isHovered := float64(mouseX) >= btn.x && float64(mouseX) <= btn.x+float64(textWidth) &&
		float64(mouseY) >= btn.y && float64(mouseY) <= btn.y+btn.size

	// Draw icon
	op := &ebiten.DrawImageOptions{}
	if isHovered {
		op.ColorScale.ScaleWithColor(color.RGBA{255, 215, 0, 255})
	}
	op.GeoM.Translate(btn.x, btn.y)
	screen.DrawImage(btn.icon, op)

	// Draw "Shop" text
	var shopTextColor color.Color = color.White
	if isHovered {
		shopTextColor = color.RGBA{255, 215, 0, 255}
	}
	text.Draw(screen, "Shop", assets.FontSmall, int(btn.x)+40, int(btn.y)+22, shopTextColor)
}

func (m *Menu) Update() {
	if m.cooldown > 0 {
		m.cooldown--
		return
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		m.readyToPlay = true
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mouseX, mouseY := ebiten.CursorPosition()

		// Check settings button
		btn := m.settingsButton
		if float64(mouseX) >= btn.x && float64(mouseX) <= btn.x+btn.size &&
			float64(mouseY) >= btn.y && float64(mouseY) <= btn.y+btn.size {
			m.openSettings = true
			return
		}

		// Check shop button (icon + text area)
		shopBtn := m.shopButton
		textWidth := 70.0 // Approximate width for icon + "Shop" text
		if float64(mouseX) >= shopBtn.x && float64(mouseX) <= shopBtn.x+textWidth &&
			float64(mouseY) >= shopBtn.y && float64(mouseY) <= shopBtn.y+shopBtn.size {
			m.openShop = true
			return
		}

		m.readyToPlay = true
	}

	touchIDs := inpututil.AppendJustPressedTouchIDs(nil)
	if len(touchIDs) > 0 {
		m.readyToPlay = true
	}
}

func (m *Menu) IsReady() bool {
	return m.readyToPlay
}

func (m *Menu) ShouldOpenSettings() bool {
	if m.openSettings {
		m.openSettings = false
		return true
	}
	return false
}

func (m *Menu) ShouldOpenShop() bool {
	if m.openShop {
		m.openShop = false
		return true
	}
	return false
}

func (m *Menu) Reset() {
	m.readyToPlay = false
	m.openSettings = false
	m.openShop = false
	m.cooldown = 15
}

func (m *Menu) Layout(outsideWidth, outsideHeight int) (int, int) {
	return config.ScreenWidth, config.ScreenHeight
}

func (m *Menu) SetScores(highScore, lastScore int) {
	m.highScore = highScore
	m.lastScore = lastScore
}
