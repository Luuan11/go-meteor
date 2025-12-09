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
	position      systems.Vector
	velocity      systems.Vector
	health        int
	size          float64
	parentBoss    *Boss
	shootCooldown int
	targetPlayer  systems.Vector
	speed         float64
	side          float64
	offsetX       float64
}

func NewMinion(boss *Boss, offsetAngle float64) *Minion {
	side := -1.0
	offsetX := 60.0

	if offsetAngle == 1 {
		side = 1.0
	} else if offsetAngle == 2 {
		side = 0.0
		offsetX = 0.0
	}

	spawnX := boss.position.X + (side * offsetX)
	spawnY := boss.position.Y

	return &Minion{
		position:      systems.Vector{X: spawnX, Y: spawnY},
		velocity:      systems.Vector{X: 0, Y: 0},
		health:        config.BossMinionHealth,
		size:          config.BossMinionSize,
		parentBoss:    boss,
		shootCooldown: 0,
		speed:         1.2,
		side:          side,
		offsetX:       offsetX,
	}
}

func (m *Minion) Update() {
	if m.parentBoss == nil {
		return
	}
	targetX := m.parentBoss.position.X + (m.side * m.offsetX)
	dx := targetX - m.position.X
	m.velocity.X = dx * 0.15

	if m.targetPlayer.Y != 0 {
		dy := m.targetPlayer.Y - m.position.Y

		if dy > 0 && (m.position.Y-m.parentBoss.position.Y) < 150 {
			m.velocity.Y = (dy / math.Abs(dy)) * m.speed
		} else if dy < -20 {
			m.velocity.Y = -m.speed * 0.8
		} else {
			m.velocity.Y *= 0.9
		}
	}

	m.position.X += m.velocity.X
	m.position.Y += m.velocity.Y

	if m.position.X < 20 {
		m.position.X = 20
	}
	if m.position.X > config.ScreenWidth-20 {
		m.position.X = config.ScreenWidth - 20
	}

	if m.position.Y < m.parentBoss.position.Y-20 {
		m.position.Y = m.parentBoss.position.Y - 20
	}

	if m.shootCooldown > 0 {
		m.shootCooldown--
	}
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

func (m *Minion) SetTarget(target systems.Vector) {
	m.targetPlayer = target
}

func (m *Minion) CanShoot() bool {
	return m.shootCooldown <= 0
}

func (m *Minion) Shoot() {
	m.shootCooldown = 90
}
