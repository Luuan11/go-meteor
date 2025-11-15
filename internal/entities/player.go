package entities

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"go-meteor/internal/config"
	"go-meteor/internal/systems"
	assets "go-meteor/src/pkg"
)

// GameInterface defines the methods that Player needs from Game
type GameInterface interface {
	AddLaser(laser *Laser)
	GetSuperPowerActive() bool
	ResetCombo()
}

type Player struct {
	game GameInterface

	position systems.Vector
	sprite   *ebiten.Image

	shootCooldown      *systems.Timer
	invincibilityTimer *systems.Timer
	shieldTimer        *systems.Timer
	isInvincible       bool
	hasShield          bool
	lives              int
}

func NewPlayer(game GameInterface) *Player {
	sprite := assets.PlayerSprite

	bounds := sprite.Bounds()
	halfW := float64(bounds.Dx()) / 2

	pos := systems.Vector{
		X: (config.ScreenWidth / 2) - halfW,
		Y: (config.ScreenHeight) - 170,
	}

	return &Player{
		game:               game,
		position:           pos,
		sprite:             sprite,
		shootCooldown:      systems.NewTimer(config.PlayerShootCooldown),
		invincibilityTimer: systems.NewTimer(config.InvincibilityTime),
		shieldTimer:        systems.NewTimer(config.ShieldTime),
		isInvincible:       false,
		hasShield:          false,
		lives:              config.InitialLives,
	}
}

func (p *Player) MoveLeft() {
	p.position.X -= config.PlayerSpeed
	if p.position.X < 0 {
		p.position.X = 0
	}
}

func (p *Player) MoveRight() {
	p.position.X += config.PlayerSpeed
	bounds := p.sprite.Bounds()
	maxX := float64(config.ScreenWidth) - float64(bounds.Dx())
	if p.position.X > maxX {
		p.position.X = maxX
	}
}

func (p *Player) MoveUp() {
	p.position.Y -= config.PlayerSpeed
	if p.position.Y < 0 {
		p.position.Y = 0
	}
}

func (p *Player) MoveDown() {
	p.position.Y += config.PlayerSpeed
	bounds := p.sprite.Bounds()
	maxY := float64(config.ScreenHeight) - float64(bounds.Dy())
	if p.position.Y > maxY {
		p.position.Y = maxY
	}
}

func (p *Player) Shoot() {
	if !p.shootCooldown.IsReady() {
		return
	}

	p.shootCooldown.Reset()

	bounds := p.sprite.Bounds()
	halfW := float64(bounds.Dx()) / 2
	halfH := float64(bounds.Dy()) / 2

	spawnPos := systems.Vector{
		X: p.position.X + halfW,
		Y: p.position.Y - halfH/2,
	}

	superPowerActive := p.game.GetSuperPowerActive()
	bullet := NewLaser(spawnPos, superPowerActive)
	p.game.AddLaser(bullet)

	if superPowerActive {
		spawnLeftPos := systems.Vector{
			X: p.position.X - halfW,
			Y: p.position.Y,
		}
		spawnRightPos := systems.Vector{
			X: p.position.X + halfW*3,
			Y: p.position.Y,
		}

		bulletLeft := NewLaser(spawnLeftPos, true)
		bulletRight := NewLaser(spawnRightPos, true)
		p.game.AddLaser(bulletLeft)
		p.game.AddLaser(bulletRight)
	}

	assets.PlayShootSound()
}

func (p *Player) Update() {
	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		p.MoveLeft()
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		p.MoveRight()
	}
	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		p.MoveUp()
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		p.MoveDown()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		p.Shoot()
	}

	p.shootCooldown.Update()
}

func (p *Player) UpdateTimers() {
	if p.isInvincible {
		p.invincibilityTimer.Update()
		if p.invincibilityTimer.IsReady() {
			p.isInvincible = false
		}
	}

	if p.hasShield {
		p.shieldTimer.Update()
		if p.shieldTimer.IsReady() {
			p.hasShield = false
		}
	}
}

func (p *Player) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}

	if p.hasShield {
		if (p.shieldTimer.CurrentTicks()/5)%2 == 0 {
			op.ColorM.Scale(0.5, 0.7, 1, 1)
		} else {
			op.ColorM.Scale(0.6, 0.8, 1, 1)
		}
	} else if p.isInvincible {
		if (p.invincibilityTimer.CurrentTicks()/5)%2 == 0 {
			op.ColorM.Scale(1, 0.5, 0.5, 0.7)
		}
	}

	op.GeoM.Translate(p.position.X, p.position.Y)
	screen.DrawImage(p.sprite, op)
}

func (p *Player) TakeDamage() bool {
	if p.hasShield {
		assets.PlayPowerUpSound()
		return false
	}

	if p.isInvincible {
		return false
	}

	p.lives--
	if p.lives <= 0 {
		return true
	}

	p.isInvincible = true
	p.invincibilityTimer.Reset()
	p.game.ResetCombo()

	assets.PlayDamageSound()

	return false
}

func (p *Player) Collider() systems.Rect {
	bounds := p.sprite.Bounds()

	return systems.NewRect(
		p.position.X,
		p.position.Y,
		float64(bounds.Dx()),
		float64(bounds.Dy()),
	)
}

func (p *Player) GetLives() int {
	return p.lives
}

func (p *Player) Heal() {
	if p.lives < config.InitialLives {
		p.lives++
	}
}

func (p *Player) ActivateShield() {
	p.hasShield = true
	p.shieldTimer.Reset()
}

func (p *Player) HasShield() bool {
	return p.hasShield
}

func (p *Player) ShieldProgress() float64 {
	return p.shieldTimer.Progress()
}
