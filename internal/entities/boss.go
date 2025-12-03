package entities

import (
	"go-meteor/internal/config"
	"go-meteor/internal/systems"
	assets "go-meteor/src/pkg"
	"image/color"
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
	bossType      config.BossType
	minions       []*Minion
	damageFlash   int
	spawnTime     time.Time
	damageTaken   int
	playerRef     systems.Vector
	trackingDelay float64
}

func NewBoss(bossType config.BossType) *Boss {
	var health int
	var speed float64
	var shootCooldown time.Duration
	var size float64

	switch bossType {
	case config.BossTank:
		health = config.BossTankHealth
		speed = config.BossTankSpeed
		shootCooldown = config.BossTankShootCooldown
		size = 90
	case config.BossSniper:
		health = config.BossSniperHealth
		speed = config.BossSniperSpeed
		shootCooldown = config.BossSniperShootCooldown
		size = 70
	case config.BossSwarm:
		health = config.BossSwarmHealth
		speed = config.BossSwarmSpeed
		shootCooldown = config.BossSwarmShootCooldown
		size = 80
	}

	boss := &Boss{
		position: systems.Vector{
			X: config.ScreenWidth / 2,
			Y: -100,
		},
		velocity: systems.Vector{
			X: 0,
			Y: speed,
		},
		health:        health,
		maxHealth:     health,
		shootCooldown: shootCooldown,
		lastShot:      time.Now(),
		movePattern:   0,
		patternTime:   0,
		size:          size,
		sprite:        assets.BossSprite,
		bossType:      bossType,
		damageFlash:   0,
		spawnTime:     time.Now(),
		damageTaken:   0,
		trackingDelay: 0,
	}

	if bossType == config.BossSwarm {
		boss.minions = make([]*Minion, config.BossMinionCount)
		for i := 0; i < config.BossMinionCount; i++ {
			angle := float64(i) * (2 * 3.14159 / float64(config.BossMinionCount))
			boss.minions[i] = NewMinion(boss, angle)
		}
	}

	return boss
}

func (b *Boss) Update() {
	b.patternTime += 0.05

	if b.damageFlash > 0 {
		b.damageFlash--
	}

	if b.position.Y < 100 {
		b.position.Y += b.velocity.Y
	} else {
		switch b.bossType {
		case config.BossTank:
			b.updateTankMovement()
		case config.BossSniper:
			b.updateSniperMovement()
		case config.BossSwarm:
			b.updateSwarmMovement()
		}

		if b.patternTime > 100 {
			b.movePattern = (b.movePattern + 1) % 4
			b.patternTime = 0
		}
	}

	if b.position.X < b.size {
		b.position.X = b.size
	}
	if b.position.X > config.ScreenWidth-b.size {
		b.position.X = config.ScreenWidth - b.size
	}

	for _, minion := range b.minions {
		if minion != nil {
			minion.Update()
		}
	}
}

func (b *Boss) updateTankMovement() {
	switch b.movePattern {
	case 0:
		b.position.X += math.Sin(b.patternTime) * 2
	case 1:
		if b.position.X < config.ScreenWidth/2 {
			b.position.X += 1.5
		} else {
			b.position.X -= 1.5
		}
	case 2:
		b.position.X += math.Cos(b.patternTime*0.5) * 2.5
	case 3:
		if int(b.patternTime*10)%40 < 20 {
			b.position.X += 2
		} else {
			b.position.X -= 2
		}
	}
}

func (b *Boss) updateSniperMovement() {
	switch b.movePattern {
	case 0:
		b.position.X += math.Sin(b.patternTime*1.5) * 5
	case 1:
		if b.position.X < 150 {
			b.position.X += 6
		} else if b.position.X > config.ScreenWidth-150 {
			b.position.X -= 6
		} else {
			if int(b.patternTime)%2 == 0 {
				b.position.X += 6
			} else {
				b.position.X -= 6
			}
		}
	case 2:
		b.trackingDelay = b.trackingDelay*0.95 + b.playerRef.X*0.05
		targetX := b.trackingDelay
		if b.position.X < targetX-5 {
			b.position.X += 4
		} else if b.position.X > targetX+5 {
			b.position.X -= 4
		}
	case 3:
		radius := 150.0
		centerX := float64(config.ScreenWidth / 2)
		b.position.X = centerX + math.Cos(b.patternTime*0.8)*radius
	}
}

func (b *Boss) updateSwarmMovement() {
	switch b.movePattern {
	case 0:
		b.position.X += math.Sin(b.patternTime) * 3.5
	case 1:
		radius := 120.0
		centerX := float64(config.ScreenWidth / 2)
		b.position.X = centerX + math.Cos(b.patternTime)*radius
	case 2:
		b.trackingDelay = b.trackingDelay*0.9 + b.playerRef.X*0.1
		targetX := b.trackingDelay
		if b.position.X < targetX-10 {
			b.position.X += 3
		} else if b.position.X > targetX+10 {
			b.position.X -= 3
		}
	case 3:
		if int(b.patternTime*10)%30 < 15 {
			b.position.X += 4
		} else {
			b.position.X -= 4
		}
	}
}

func (b *Boss) Draw(screen *ebiten.Image) {
	for _, minion := range b.minions {
		if minion != nil {
			minion.Draw(screen)
		}
	}

	op := &ebiten.DrawImageOptions{}

	bounds := b.sprite.Bounds()
	scale := b.size * 2 / float64(bounds.Dx())

	op.GeoM.Translate(-float64(bounds.Dx())/2, -float64(bounds.Dy())/2)
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(b.position.X, b.position.Y)

	if b.damageFlash > 0 {
		op.ColorScale.ScaleWithColor(color.RGBA{255, 100, 100, 255})
	}

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
	b.damageTaken += damage
	b.damageFlash = 10
	return b.health <= 0
}

func (b *Boss) GetHealth() int {
	return b.health
}

func (b *Boss) GetMaxHealth() int {
	return b.maxHealth
}

func (b *Boss) GetBossType() config.BossType {
	return b.bossType
}

func (b *Boss) GetMinions() []*Minion {
	return b.minions
}

func (b *Boss) RemoveMinion(index int) {
	if index >= 0 && index < len(b.minions) {
		b.minions[index] = nil
	}
}

func (b *Boss) SetPlayerPosition(pos systems.Vector) {
	b.playerRef = pos
}

func (b *Boss) GetSpawnTime() time.Time {
	return b.spawnTime
}

func (b *Boss) GetDamageTaken() int {
	return b.damageTaken
}

func (b *Boss) IsOutOfScreen() bool {
	return b.position.Y > config.ScreenHeight+100
}
