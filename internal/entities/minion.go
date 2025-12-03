package entities

import (
	"go-meteor/internal/config"
	"go-meteor/internal/systems"
	"math"

	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Minion struct {
	position    systems.Vector
	offset      float64
	offsetAngle float64
	health      int
	size        float64
	parentBoss  *Boss
}

func NewMinion(boss *Boss, offsetAngle float64) *Minion {
	return &Minion{
		position:    boss.position,
		offset:      40,
		offsetAngle: offsetAngle,
		health:      config.BossMinionHealth,
		size:        15,
		parentBoss:  boss,
	}
}

func (m *Minion) Update() {
	m.offsetAngle += 0.05

	m.position.X = m.parentBoss.position.X + math.Cos(m.offsetAngle)*m.offset
	m.position.Y = m.parentBoss.position.Y + math.Sin(m.offsetAngle)*m.offset
}

func (m *Minion) Draw(screen *ebiten.Image) {
	vector.DrawFilledCircle(screen, float32(m.position.X), float32(m.position.Y), float32(m.size), color.RGBA{150, 50, 200, 255}, false)
	vector.DrawFilledCircle(screen, float32(m.position.X), float32(m.position.Y), float32(m.size-3), color.RGBA{200, 100, 255, 200}, false)
}

func (m *Minion) Collider() systems.Rect {
	return systems.Rect{
		X:      m.position.X,
		Y:      m.position.Y,
		Width:  m.size * 2,
		Height: m.size * 2,
	}
}

func (m *Minion) GetPosition() systems.Vector {
	return m.position
}

func (m *Minion) TakeDamage(damage int) bool {
	m.health -= damage
	return m.health <= 0
}

func (m *Minion) GetHealth() int {
	return m.health
}
