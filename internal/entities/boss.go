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
	direction     float64
}

func NewBoss(bossType config.BossType) *Boss {
	var health int
	var speed float64
	var shootCooldown time.Duration
	var size float64
	var sprite *ebiten.Image

	switch bossType {
	case config.BossTank:
		health = config.BossTankHealth
		speed = config.BossTankSpeed
		shootCooldown = config.BossTankShootCooldown
		size = 120
		sprite = assets.BossTankSprite
	case config.BossSniper:
		health = config.BossSniperHealth
		speed = config.BossSniperSpeed
		shootCooldown = config.BossSniperShootCooldown
		size = 90
		sprite = assets.BossSniperSprite
	case config.BossSwarm:
		health = config.BossSwarmHealth
		speed = config.BossSwarmSpeed
		shootCooldown = config.BossSwarmShootCooldown
		size = 100
		sprite = assets.BossSwarmSprite
	}

	startX := float64(config.ScreenWidth) / 4.0
	direction := 1.0
	if time.Now().UnixNano()%2 == 0 {
		startX = float64(config.ScreenWidth) * 3.0 / 4.0
		direction = -1.0
	}

	boss := &Boss{
		position: systems.Vector{
			X: startX,
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
		movePattern:   int(time.Now().UnixNano() % 4),
		patternTime:   0,
		size:          size,
		sprite:        sprite,
		bossType:      bossType,
		damageFlash:   0,
		spawnTime:     time.Now(),
		damageTaken:   0,
		trackingDelay: startX,
		direction:     direction,
	}

	minionCount := config.BossMinionCount
	if bossType == config.BossSwarm {
		minionCount = config.BossSwarmMinionCount
	}

	boss.minions = make([]*Minion, minionCount)
	for i := 0; i < minionCount; i++ {
		angle := float64(i)
		boss.minions[i] = NewMinion(boss, angle)
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

	if b.position.X < 50 {
		b.position.X = 50
	}
	if b.position.X > config.ScreenWidth-50 {
		b.position.X = config.ScreenWidth - 50
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
		b.position.X += b.direction * 1.5
		if b.position.X < 50 || b.position.X > config.ScreenWidth-50 {
			b.direction *= -1
		}
	case 1:
		b.position.X += b.direction * 2
		if b.position.X < 50 || b.position.X > config.ScreenWidth-50 {
			b.direction *= -1
		}
	case 2:
		b.position.X += b.direction * 2.5
		b.position.X += math.Sin(b.patternTime*0.8) * 1.5
		if b.position.X < 50 || b.position.X > config.ScreenWidth-50 {
			b.direction *= -1
		}
	case 3:
		if int(b.patternTime*10)%40 < 20 {
			b.position.X += 2.5
		} else {
			b.position.X -= 2.5
		}
	}
}

func (b *Boss) updateSniperMovement() {
	switch b.movePattern {
	case 0:
		b.position.X += b.direction * 3
		b.position.X += math.Sin(b.patternTime*1.5) * 1.5
		if b.position.X < 50 || b.position.X > config.ScreenWidth-50 {
			b.direction *= -1
		}
	case 1:
		b.position.X += b.direction * 5
		if b.position.X < 50 || b.position.X > config.ScreenWidth-50 {
			b.direction *= -1
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
		radius := 300.0
		centerX := float64(config.ScreenWidth / 2)
		offset := math.Cos(b.patternTime*0.8) * radius
		if centerX+offset < 60 {
			offset = 60 - centerX
		} else if centerX+offset > config.ScreenWidth-60 {
			offset = config.ScreenWidth - 60 - centerX
		}
		b.position.X = centerX + offset
	}
}

func (b *Boss) updateSwarmMovement() {
	switch b.movePattern {
	case 0:
		b.position.X += b.direction * 2.5
		b.position.X += math.Sin(b.patternTime*0.9) * 1
		if b.position.X < 50 || b.position.X > config.ScreenWidth-50 {
			b.direction *= -1
		}
	case 1:
		radius := 280.0
		centerX := float64(config.ScreenWidth / 2)
		offset := math.Cos(b.patternTime*0.7) * radius
		if centerX+offset < 70 {
			offset = 70 - centerX
		} else if centerX+offset > config.ScreenWidth-70 {
			offset = config.ScreenWidth - 70 - centerX
		}
		b.position.X = centerX + offset
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
	if b == nil {
		return nil
	}
	return b.minions
}

func (b *Boss) RemoveMinion(index int) {
	if b == nil || b.minions == nil {
		return
	}
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
