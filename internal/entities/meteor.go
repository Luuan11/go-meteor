package entities

import (
	"go-meteor/internal/config"
	"go-meteor/internal/systems"
	assets "go-meteor/src/pkg"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
)

type Meteor struct {
	position      systems.Vector
	rotation      float64
	movement      systems.Vector
	rotationSpeed float64
	sprite        *ebiten.Image
}

func NewMeteor(speedMultiplier float64) *Meteor {
	pos := systems.Vector{
		X: rand.Float64() * config.ScreenWidth,
		Y: -100,
	}

	velocity := config.MeteorMinSpeed + rand.Float64()*(config.MeteorMaxSpeed-config.MeteorMinSpeed)
	velocity *= speedMultiplier

	movement := systems.Vector{
		X: 0,
		Y: velocity,
	}

	sprite := assets.MeteorSprites[rand.Intn(len(assets.MeteorSprites))]

	m := &Meteor{
		position:      pos,
		movement:      movement,
		rotationSpeed: config.MeteorRotationMin + rand.Float64()*(config.MeteorRotationMax-config.MeteorRotationMin),
		sprite:        sprite,
	}
	return m
}

func (m *Meteor) Reset(speedMultiplier float64) {
	m.position = systems.Vector{
		X: rand.Float64() * config.ScreenWidth,
		Y: -100,
	}

	velocity := config.MeteorMinSpeed + rand.Float64()*(config.MeteorMaxSpeed-config.MeteorMinSpeed)
	velocity *= speedMultiplier

	m.movement = systems.Vector{
		X: 0,
		Y: velocity,
	}

	m.rotation = 0
	m.rotationSpeed = config.MeteorRotationMin + rand.Float64()*(config.MeteorRotationMax-config.MeteorRotationMin)
	m.sprite = assets.MeteorSprites[rand.Intn(len(assets.MeteorSprites))]
}

func (m *Meteor) Update() {
	m.position.X += m.movement.X
	m.position.Y += m.movement.Y
	m.rotation += m.rotationSpeed
}

func (m *Meteor) Draw(screen *ebiten.Image) {
	bounds := m.sprite.Bounds()
	halfW := float64(bounds.Dx()) / 2
	halfH := float64(bounds.Dy()) / 2

	op := &ebiten.DrawImageOptions{}

	op.GeoM.Translate(-halfW, -halfH)
	op.GeoM.Rotate(m.rotation)
	op.GeoM.Translate(halfW, halfH)

	op.GeoM.Translate(m.position.X, m.position.Y)

	screen.DrawImage(m.sprite, op)
}

func (m *Meteor) Collider() systems.Rect {
	bounds := m.sprite.Bounds()

	return systems.NewRect(
		m.position.X,
		m.position.Y,
		float64(bounds.Dx()),
		float64(bounds.Dy()),
	)
}

func (m *Meteor) IsOutOfScreen() bool {
	return m.position.Y > config.ScreenHeight+100
}

func (m *Meteor) GetPosition() systems.Vector {
	return m.position
}

func (m *Meteor) GetMovement() systems.Vector {
	return m.movement
}

func (m *Meteor) SetMovementY(y float64) {
	m.movement.Y = y
}

func (m *Meteor) ApplySlowMotion(factor float64) {
	m.movement.X *= factor
	m.movement.Y *= factor
}

func (m *Meteor) RestoreSpeed() {
	m.movement.X /= config.SlowMotionFactor
	m.movement.Y /= config.SlowMotionFactor
}
