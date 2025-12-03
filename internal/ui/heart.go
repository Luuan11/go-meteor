package ui

import (
	assets "go-meteor/src/pkg"

	"github.com/hajimehoshi/ebiten/v2"
)

func DrawHeart(screen *ebiten.Image, x, y int, filled bool) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	
	if filled {
		screen.DrawImage(assets.HeartFilledSprite, op)
	} else {
		screen.DrawImage(assets.HeartEmptySprite, op)
	}
}
