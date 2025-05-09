package menu

import (
	"go-meteor/src/pkg"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
)

const (
	screenWidth  = 800
	screenHeight = 600
)

type Menu struct {
	readyToPlay bool
}

func NewMenu() *Menu {

	return &Menu{
		readyToPlay: false,
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
	if ebiten.IsKeyPressed(ebiten.KeyEnter) {
		m.readyToPlay = true
	}

	var touchIDs []ebiten.TouchID
    touchIDs = ebiten.AppendTouchIDs(touchIDs)
    if len(touchIDs) > 0 {
        m.readyToPlay = true
    }
}

func (m *Menu) IsReady() bool {
	return m.readyToPlay
}
func (m *Menu) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}