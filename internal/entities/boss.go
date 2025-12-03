package entities

import (
	"go-meteor/internal/config"
	"go-meteor/internal/systems"
	assets "go-meteor/src/pkg"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

type Boss struct {
	position      systems.Vector
	velocity      systems.Vector
	health        int
	maxHealth     int
	shootCooldown time.Duration
	lastShot      time.Time
	movePattern   int
	patternTime   float64
	size          float64
	sprite        *ebiten.Image
}

func NewBoss() *Boss {
	return &Boss{
		position: systems.Vector{
			X: config.ScreenWidth / 2,
			Y: -100,
		},
		velocity: systems.Vector{
			X: 0,
			Y: config.BossSpeed,
		},
		health:        config.BossHealth,
		maxHealth:     config.BossHealth,
		shootCooldown: config.BossShootCooldown,
		lastShot:      time.Now(),
		movePattern:   0,
		patternTime:   0,
		size:          80,
		sprite:        assets.BossSprite,
	}
}

func (b *Boss) Update() {
	b.patternTime += 0.05

	if b.position.Y < 100 {
		b.position.Y += b.velocity.Y
	} else {
		switch b.movePattern {
		case 0:
			b.position.X += math.Sin(b.patternTime) * 4
		case 1:
			if b.position.X < 100 {
				b.position.X += 3
			} else if b.position.X > config.ScreenWidth-100 {
				b.position.X -= 3
			} else {
				if int(b.patternTime)%2 == 0 {
					b.position.X += 3
				} else {
					b.position.X -= 3
				}
			}
		case 2:
			b.position.X += math.Cos(b.patternTime*0.8) * 5
		}

		if b.patternTime > 100 {
			b.movePattern = (b.movePattern + 1) % 3
			b.patternTime = 0
		}
	}

	if b.position.X < 80 {
		b.position.X = 80
	}
	if b.position.X > config.ScreenWidth-80 {
		b.position.X = config.ScreenWidth - 80
	}
}

func (b *Boss) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}

	bounds := b.sprite.Bounds()
	scale := b.size * 2 / float64(bounds.Dx())

	op.GeoM.Translate(-float64(bounds.Dx())/2, -float64(bounds.Dy())/2)
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(b.position.X, b.position.Y)

	screen.DrawImage(b.sprite, op)
}

func (b *Boss) Collider() systems.Rect {
	return systems.Rect{
		X:      b.position.X,
		Y:      b.position.Y,
		Width:  b.size,
		Height: b.size,
	}
}

func (b *Boss) GetPosition() systems.Vector {
	return b.position
}

func (b *Boss) CanShoot() bool {
	return time.Since(b.lastShot) >= b.shootCooldown
}

func (b *Boss) Shoot() {
	b.lastShot = time.Now()
}

func (b *Boss) TakeDamage(damage int) bool {
	b.health -= damage
	return b.health <= 0
}

func (b *Boss) GetHealth() int {
	return b.health
}

func (b *Boss) GetMaxHealth() int {
	return b.maxHealth
}

func (b *Boss) IsOutOfScreen() bool {
	return b.position.Y > config.ScreenHeight+100
}
