package game

import (
	"go-meteor/assets"

	"github.com/hajimehoshi/ebiten/v2"
)

type Player struct {
	image *ebiten.Image
}

func NewPlayer() *Player {
	image := assets.PlayerBody
	return &Player{
		image: image,
	}
}

func (p *Player) Update() error {
	return nil
}

func (p *Player) Draw(screen *ebiten.Image) {
	permission := &ebiten.DrawImageOptions{}

	permission.GeoM.Translate(100, 100)
	screen.DrawImage(p.image, permission)
}
