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
	isLaserBeam   bool
	damage        int
}

func NewLaser(pos systems.Vector, isSuperPower bool, isLaserBeam bool) *Laser {
	sprite := assets.LaserSprite
	speed := config.LaserSpeed
	damage := 1

	if isLaserBeam {
		sprite = assets.LaserBeamSprite
		speed = config.SuperLaserSpeed
		damage = 3
	} else if isSuperPower {
		sprite = assets.SuperPowerSprite
		speed = config.SuperLaserSpeed
		damage = 2
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
		isLaserBeam:   isLaserBeam,
		damage:        damage,
	}

	return b
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

func (l *Laser) IsLaserBeam() bool {
	return l.isLaserBeam
}

func (l *Laser) GetDamage() int {
	return l.damage
}

func (l *Laser) IsOutOfScreen() bool {
	return l.position.Y < -100
}
