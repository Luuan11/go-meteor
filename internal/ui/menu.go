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
	readyToPlay bool
	cooldown    int
	highScore   int
	lastScore   int
}

func NewMenu() *Menu {

	return &Menu{
		readyToPlay: false,
		cooldown:    0,
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

func (m *Menu) Reset() {
	m.readyToPlay = false
	m.cooldown = 15 // 15 frames de cooldown (~0.25 segundos)
}

func (m *Menu) Layout(outsideWidth, outsideHeight int) (int, int) {
	return config.ScreenWidth, config.ScreenHeight
}

func (m *Menu) SetScores(highScore, lastScore int) {
	m.highScore = highScore
	m.lastScore = lastScore
}
