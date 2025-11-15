package ui

import (
	assets "go-meteor/src/pkg"

	"github.com/hajimehoshi/ebiten/v2"
)

func DrawPauseIcon(screen *ebiten.Image, x, y int) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(assets.PauseIcon, op)
}
