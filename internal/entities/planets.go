package entities

import (
	"go-meteor/internal/config"
	"go-meteor/internal/systems"
	assets "go-meteor/src/pkg"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
)

type Planet struct {
	position      systems.Vector
	rotation      float64
	movement      systems.Vector
	rotationSpeed float64
	sprite        *ebiten.Image
}

func NewPlanet() *Planet {
	pos := systems.Vector{
		X: rand.Float64() * config.ScreenWidth,
		Y: -500,
	}

	velocity := float64(2)

	movement := systems.Vector{
		X: 0,
		Y: velocity,
	}

	sprite := assets.PlanetsSprites[rand.Intn(len(assets.PlanetsSprites))]

	m := &Planet{
		position: pos,
		movement: movement,
		sprite:   sprite,
	}
	return m
}

func (m *Planet) IsOutOfScreen() bool {
	return m.position.Y > config.ScreenHeight+500
}

func (m *Planet) Update() {
	m.position.X += m.movement.X
	m.position.Y += m.movement.Y
	m.rotation += m.rotationSpeed
}

func (m *Planet) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(m.position.X, m.position.Y)
	screen.DrawImage(m.sprite, op)
}
