package ui

import (
	"go-meteor/internal/config"
	assets "go-meteor/src/pkg"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
)

type Menu struct {
	readyToPlay bool
	cooldown    int
}

func NewMenu() *Menu {

	return &Menu{
		readyToPlay: false,
		cooldown:    0,
	}
}

func (m *Menu) Draw(screen *ebiten.Image) {

	text.Draw(screen, "Space GO", assets.ScoreFont, 270, 300, color.White)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(315, 150)
	screen.DrawImage(assets.GopherPlayer, op)

	text.Draw(screen, "Press ENTER to start", assets.FontUi, 100, 400, color.White)
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
