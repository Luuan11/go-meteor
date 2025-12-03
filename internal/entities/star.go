package entities

import (
	"go-meteor/internal/config"
	"go-meteor/internal/systems"
	"image/color"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Star struct {
	position systems.Vector
	movement systems.Vector
	size     float32
}

func NewStar() *Star {
	pos := systems.Vector{
		X: rand.Float64() * config.ScreenWidth,
		Y: -10,
	}

	velocity := 2.0 + rand.Float64()*2.0

	movement := systems.Vector{
		X: 0,
		Y: velocity,
	}

	size := 1.0 + rand.Float32()*1.5

	m := &Star{
		position: pos,
		movement: movement,
		size:     size,
	}
	return m
}

func (m *Star) IsOutOfScreen() bool {
	return m.position.Y > config.ScreenHeight+100
}

func (m *Star) Update() {
	m.position.X += m.movement.X
	m.position.Y += m.movement.Y
}

func (m *Star) Draw(screen *ebiten.Image) {
	vector.DrawFilledCircle(screen, float32(m.position.X), float32(m.position.Y), m.size, color.RGBA{255, 255, 100, 255}, false)
}
