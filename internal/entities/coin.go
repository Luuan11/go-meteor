package entities

import (
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"

	"go-meteor/internal/config"
	"go-meteor/internal/systems"
	assets "go-meteor/src/pkg"
)

type Coin struct {
	position  systems.Vector
	movement  systems.Vector
	sprite    *ebiten.Image
	value     int
	collected bool
	targetX   float64
	targetY   float64
	speed     float64
}

const (
	CoinDropChance = 0.20
	CoinSpeed      = 2.0
	CoinValue      = 1
	BossCoinValue  = 10
	CoinSize       = 20
)

func NewCoin(x, y float64, value int) *Coin {
	return &Coin{
		position:  systems.Vector{X: x, Y: y},
		movement:  systems.Vector{X: 0, Y: CoinSpeed},
		sprite:    assets.CoinSprite,
		value:     value,
		collected: false,
		speed:     8.0,
	}
}

func NewCoinFromMeteor(meteor *Meteor) *Coin {
	bounds := meteor.sprite.Bounds()
	centerX := meteor.position.X + float64(bounds.Dx())/2 - CoinSize/2
	centerY := meteor.position.Y + float64(bounds.Dy())/2 - CoinSize/2
	return NewCoin(centerX, centerY, CoinValue)
}

func NewCoinFromBoss(x, y float64) *Coin {
	return NewCoin(x, y, BossCoinValue)
}

func (c *Coin) Update() {
	if c.collected {
		dx := c.targetX - c.position.X
		dy := c.targetY - c.position.Y

		dist := dx*dx + dy*dy
		if dist > 1 {
			length := 1.0
			if dist > 0 {
				length = 1.0 / (dist * 0.01)
				if length < 0.01 {
					length = 0.01
				}
			}
			c.position.X += dx * c.speed * length
			c.position.Y += dy * c.speed * length
			c.speed += 0.5
		}
	} else {
		c.position.Y += c.movement.Y
	}
}

func (c *Coin) Collect(targetX, targetY float64) {
	c.collected = true
	c.targetX = targetX
	c.targetY = targetY
}

func (c *Coin) IsCollected() bool {
	return c.collected
}

func (c *Coin) HasReachedTarget() bool {
	if !c.collected {
		return false
	}
	dx := c.targetX - c.position.X
	dy := c.targetY - c.position.Y
	return dx*dx+dy*dy < 100
}

func (c *Coin) Draw(screen *ebiten.Image) {
	if c.sprite == nil {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(c.position.X, c.position.Y)
	screen.DrawImage(c.sprite, op)
}

func (c *Coin) IsOffScreen() bool {
	return c.position.Y > config.ScreenHeight+50
}

func (c *Coin) GetBounds() (float64, float64, float64, float64) {
	return c.position.X, c.position.Y, CoinSize, CoinSize
}

func (c *Coin) GetValue() int {
	return c.value
}

func ShouldDropCoin() bool {
	return rand.Float64() < CoinDropChance
}
