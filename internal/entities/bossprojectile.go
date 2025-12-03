package entities

import (
	"go-meteor/internal/config"
	"go-meteor/internal/systems"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type BossProjectile struct {
	position systems.Vector
	velocity systems.Vector
	size     float64
}

func NewBossProjectile(x, y float64) *BossProjectile {
	return &BossProjectile{
		position: systems.Vector{X: x, Y: y},
		velocity: systems.Vector{X: 0, Y: config.BossProjectileSpeed},
		size:     10,
	}
}

func (bp *BossProjectile) Reset(x, y float64) {
	bp.position.X = x
	bp.position.Y = y
	bp.velocity.X = 0
	bp.velocity.Y = config.BossProjectileSpeed
	bp.size = 10
}

func (bp *BossProjectile) Update() {
	bp.position.X += bp.velocity.X
	bp.position.Y += bp.velocity.Y
}

func (bp *BossProjectile) Draw(screen *ebiten.Image) {
	projectileColor := color.RGBA{255, 50, 50, 255}
	vector.DrawFilledCircle(screen, float32(bp.position.X), float32(bp.position.Y), float32(bp.size), projectileColor, false)

	glowColor := color.RGBA{255, 100, 100, 100}
	vector.DrawFilledCircle(screen, float32(bp.position.X), float32(bp.position.Y), float32(bp.size+3), glowColor, false)
}

func (bp *BossProjectile) Collider() systems.Rect {
	return systems.Rect{
		X:      bp.position.X,
		Y:      bp.position.Y,
		Width:  bp.size * 2,
		Height: bp.size * 2,
	}
}

func (bp *BossProjectile) IsOutOfScreen() bool {
	return bp.position.Y > config.ScreenHeight+50 ||
		bp.position.Y < -50 ||
		bp.position.X < -50 ||
		bp.position.X > config.ScreenWidth+50
}

func (bp *BossProjectile) GetPosition() systems.Vector {
	return bp.position
}
