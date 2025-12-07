package entities

import (
	"go-meteor/internal/config"
	"go-meteor/internal/systems"
	assets "go-meteor/src/pkg"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
)

type PowerUpType int

const (
	PowerUpSuperShot PowerUpType = iota
	PowerUpHeart
	PowerUpShield
	PowerUpSlowMotion
	PowerUpLaser
	PowerUpNuke
	PowerUpExtraLife
	PowerUpMultiplier
)

type PowerUp struct {
	position  systems.Vector
	movement  systems.Vector
	sprite    *ebiten.Image
	powerType PowerUpType
}

func NewPowerUp() *PowerUp {
	pos := systems.Vector{
		X: rand.Float64() * config.ScreenWidth,
		Y: -100,
	}

	movement := systems.Vector{
		X: 0,
		Y: config.PowerUpSpeed,
	}

	powerType := PowerUpType(rand.Intn(4))
	var sprite *ebiten.Image

	switch powerType {
	case PowerUpHeart:
		sprite = assets.HeartPowerUpSprite
	case PowerUpShield:
		sprite = assets.ShieldPowerUpSprite
	case PowerUpSlowMotion:
		sprite = assets.ClockPowerUpSprite
	default:
		sprite = assets.PowerUpSprites
	}

	return &PowerUp{
		position:  pos,
		movement:  movement,
		sprite:    sprite,
		powerType: powerType,
	}
}

func (p *PowerUp) Reset() {
	p.position = systems.Vector{
		X: rand.Float64() * config.ScreenWidth,
		Y: -100,
	}

	p.movement = systems.Vector{
		X: 0,
		Y: config.PowerUpSpeed,
	}

	p.powerType = PowerUpType(rand.Intn(4))

	switch p.powerType {
	case PowerUpHeart:
		p.sprite = assets.HeartPowerUpSprite
	case PowerUpShield:
		p.sprite = assets.ShieldPowerUpSprite
	case PowerUpSlowMotion:
		p.sprite = assets.ClockPowerUpSprite
	case PowerUpLaser:
		p.sprite = assets.LaserPowerUpSprite
	case PowerUpNuke:
		p.sprite = assets.NukePowerUpSprite
	case PowerUpMultiplier:
		p.sprite = assets.MultiplierPowerUpSprite
	default:
		p.sprite = assets.PowerUpSprites
	}
}

func (p *PowerUp) Update() {
	p.position.X += p.movement.X
	p.position.Y += p.movement.Y
}

func (p *PowerUp) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(p.position.X, p.position.Y)
	screen.DrawImage(p.sprite, op)
}

func (p *PowerUp) Collider() systems.Rect {
	bounds := p.sprite.Bounds()

	return systems.NewRect(
		p.position.X,
		p.position.Y,
		float64(bounds.Dx()),
		float64(bounds.Dy()),
	)
}

func (p *PowerUp) IsOutOfScreen() bool {
	return p.position.Y > config.ScreenHeight+100
}

func (p *PowerUp) GetType() PowerUpType {
	return p.powerType
}

func NewPowerUpWithType(powerType PowerUpType) *PowerUp {
	pos := systems.Vector{
		X: rand.Float64() * config.ScreenWidth,
		Y: -100,
	}

	movement := systems.Vector{
		X: 0,
		Y: config.PowerUpSpeed,
	}

	var sprite *ebiten.Image

	switch powerType {
	case PowerUpHeart:
		sprite = assets.HeartPowerUpSprite
	case PowerUpShield:
		sprite = assets.ShieldPowerUpSprite
	case PowerUpSlowMotion:
		sprite = assets.ClockPowerUpSprite
	case PowerUpLaser:
		sprite = assets.LaserPowerUpSprite
	case PowerUpNuke:
		sprite = assets.NukePowerUpSprite
	case PowerUpExtraLife:
		sprite = assets.ExtraLifePowerUpSprite
	case PowerUpMultiplier:
		sprite = assets.MultiplierPowerUpSprite
	default:
		sprite = assets.PowerUpSprites
	}

	return &PowerUp{
		position:  pos,
		movement:  movement,
		sprite:    sprite,
		powerType: powerType,
	}
}
