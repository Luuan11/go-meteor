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
	menuCooldownFrames = 15
	shopTextWidth      = 70
)

var (
	colorMenuWhite  = color.White
	colorMenuGold   = color.RGBA{255, 215, 0, 255}
	colorMenuPurple = color.RGBA{143, 47, 233, 255}
	colorMenuCredit = color.RGBA{150, 100, 255, 255}
)

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
	m.drawPlayer(screen)
	m.drawTitle(screen)
	m.drawScores(screen)
	m.drawInstructions(screen)
	m.drawCredit(screen)
	m.drawButtons(screen)
}

func (m *Menu) drawPlayer(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(0.5, 0.5)
	op.GeoM.Translate(280, 50)
	screen.DrawImage(assets.GopherPlayer, op)
}

func (m *Menu) drawTitle(screen *ebiten.Image) {
	titleText := "Space GO"
	titleBounds := text.BoundString(assets.ScoreFont, titleText)
	titleX := (config.ScreenWidth - titleBounds.Dx()) / 2
	text.Draw(screen, titleText, assets.ScoreFont, titleX, 320, colorMenuWhite)
}

func (m *Menu) drawScores(screen *ebiten.Image) {
	highScoreText := "High Score: " + fmt.Sprintf("%d", m.highScore)
	highScoreBounds := text.BoundString(assets.FontSmall, highScoreText)
	highScoreX := (config.ScreenWidth - highScoreBounds.Dx()) / 2
	text.Draw(screen, highScoreText, assets.FontSmall, highScoreX, 510, colorMenuGold)

	if m.lastScore > 0 {
		lastScoreText := "Last Score: " + fmt.Sprintf("%d", m.lastScore)
		lastScoreBounds := text.BoundString(assets.FontSmall, lastScoreText)
		lastScoreX := (config.ScreenWidth - lastScoreBounds.Dx()) / 2
		text.Draw(screen, lastScoreText, assets.FontSmall, lastScoreX, 535, colorMenuPurple)
	}
}

func (m *Menu) drawInstructions(screen *ebiten.Image) {
	instructionText := "Press ENTER to start"
	instructionBounds := text.BoundString(assets.FontUi, instructionText)
	instructionX := (config.ScreenWidth - instructionBounds.Dx()) / 2
	text.Draw(screen, instructionText, assets.FontUi, instructionX, 400, colorMenuWhite)
}

func (m *Menu) drawCredit(screen *ebiten.Image) {
	creditText := "Luuan11"
	creditBounds := text.BoundString(assets.FontSmall, creditText)
	creditX := (config.ScreenWidth - creditBounds.Dx()) / 2
	text.Draw(screen, creditText, assets.FontSmall, creditX, 565, colorMenuCredit)
}

func (m *Menu) drawButtons(screen *ebiten.Image) {
	m.drawIconButton(screen, m.settingsButton)
	m.drawShopButton(screen)
}

func (m *Menu) drawIconButton(screen *ebiten.Image, btn *IconButton) {
	op := &ebiten.DrawImageOptions{}
	if m.isButtonHovered(btn, btn.size) {
		op.ColorScale.ScaleWithColor(color.RGBA{255, 215, 0, 255})
	}
	op.GeoM.Translate(btn.x, btn.y)
	screen.DrawImage(btn.icon, op)
}

func (m *Menu) drawShopButton(screen *ebiten.Image) {
	btn := m.shopButton
	isHovered := m.isButtonHovered(btn, shopTextWidth)

	op := &ebiten.DrawImageOptions{}
	if isHovered {
		op.ColorScale.ScaleWithColor(color.RGBA{255, 215, 0, 255})
	}
	op.GeoM.Translate(btn.x, btn.y)
	screen.DrawImage(btn.icon, op)

	var shopTextColor color.Color = colorMenuWhite
	if isHovered {
		shopTextColor = color.RGBA{255, 215, 0, 255}
	}
	text.Draw(screen, "Shop", assets.FontSmall, int(btn.x)+40, int(btn.y)+22, shopTextColor)
}

func (m *Menu) isButtonHovered(btn *IconButton, width float64) bool {
	x, y := ebiten.CursorPosition()
	return float64(x) >= btn.x && float64(x) <= btn.x+width &&
		float64(y) >= btn.y && float64(y) <= btn.y+btn.size
}

func (m *Menu) Update() {
	if m.cooldown > 0 {
		m.cooldown--
		return
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		m.readyToPlay = true
	}

	m.handleMouseInput()
	m.handleTouchInput()
}

func (m *Menu) handleMouseInput() {
	if !inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return
	}

	if m.isButtonHovered(m.settingsButton, m.settingsButton.size) {
		m.openSettings = true
		return
	}

	if m.isButtonHovered(m.shopButton, shopTextWidth) {
		m.openShop = true
		return
	}

	m.readyToPlay = true
}

func (m *Menu) handleTouchInput() {
	touchIDs := inpututil.AppendJustPressedTouchIDs(nil)
	if len(touchIDs) == 0 {
		return
	}

	for _, id := range touchIDs {
		x, y := ebiten.TouchPosition(id)

		if m.isTouchOnButton(m.settingsButton, x, y, m.settingsButton.size) {
			m.openSettings = true
			return
		}

		if m.isTouchOnButton(m.shopButton, x, y, shopTextWidth) {
			m.openShop = true
			return
		}
	}

	m.readyToPlay = true
}

func (m *Menu) isTouchOnButton(btn *IconButton, x, y int, width float64) bool {
	return float64(x) >= btn.x && float64(x) <= btn.x+width &&
		float64(y) >= btn.y && float64(y) <= btn.y+btn.size
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
	m.cooldown = menuCooldownFrames
}

func (m *Menu) Layout(outsideWidth, outsideHeight int) (int, int) {
	return config.ScreenWidth, config.ScreenHeight
}

func (m *Menu) SetScores(highScore, lastScore int) {
	m.highScore = highScore
	m.lastScore = lastScore
}
