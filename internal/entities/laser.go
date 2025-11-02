package entities

import (
	"go-meteor/internal/config"
	"go-meteor/internal/systems"
	assets "go-meteor/src/pkg"

	"github.com/hajimehoshi/ebiten/v2"
)

type Laser struct {
	position      systems.Vector
	sprite        *ebiten.Image
	speed         float64
	rotation      float64
	rotationSpeed float64
	isSuperPower  bool
}

func NewLaser(pos systems.Vector, isSuperPower bool) *Laser {
	sprite := assets.LaserSprite
	speed := config.LaserSpeed

	if isSuperPower {
		sprite = assets.SuperPowerSprite
		speed = config.SuperLaserSpeed
	}

	bounds := sprite.Bounds()
	halfW := float64(bounds.Dx()) / 2
	halfH := float64(bounds.Dy()) / 2

	pos.X -= halfW
	pos.Y -= halfH

	b := &Laser{
		position:      pos,
		speed:         speed,
		rotationSpeed: config.MeteorRotationMax * 2,
		sprite:        sprite,
		isSuperPower:  isSuperPower,
	}

	return b
}

func (l *Laser) Reset(pos systems.Vector, isSuperPower bool) {
	sprite := assets.LaserSprite
	speed := config.LaserSpeed

	if isSuperPower {
		sprite = assets.SuperPowerSprite
		speed = config.SuperLaserSpeed
	}

	bounds := sprite.Bounds()
	halfW := float64(bounds.Dx()) / 2
	halfH := float64(bounds.Dy()) / 2

	pos.X -= halfW
	pos.Y -= halfH

	l.position = pos
	l.speed = speed
	l.rotation = 0
	l.rotationSpeed = config.MeteorRotationMax * 2
	l.sprite = sprite
	l.isSuperPower = isSuperPower
}

func (l *Laser) Update() {
	l.position.Y += -l.speed
	l.rotation += l.rotationSpeed
}

func (l *Laser) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}

	if l.isSuperPower && l.sprite == assets.SuperPowerSprite {
		bounds := assets.SuperPowerSprite.Bounds()
		halfW := float64(bounds.Dx()) / 2
		halfH := float64(bounds.Dy()) / 2

		op.GeoM.Translate(-halfW, -halfH)
		op.GeoM.Rotate(l.rotation)
		op.GeoM.Translate(halfW, halfH)
	}

	op.GeoM.Translate(l.position.X, l.position.Y)

	screen.DrawImage(l.sprite, op)
}

func (l *Laser) Collider() systems.Rect {
	bounds := l.sprite.Bounds()

	return systems.NewRect(
		l.position.X,
		l.position.Y,
		float64(bounds.Dx()),
		float64(bounds.Dy()),
	)
}

func (l *Laser) IsOutOfScreen() bool {
	return l.position.Y < -100
}
