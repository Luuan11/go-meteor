package effects

import (
	"go-meteor/internal/config"
	"go-meteor/internal/systems"
	"image/color"
	"math/rand"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Particle struct {
	position systems.Vector
	velocity systems.Vector
	color    color.RGBA
	life     int
	maxLife  int
}

func NewParticle(pos systems.Vector) *Particle {
	angle := rand.Float64() * 6.28318530718
	speed := rand.Float64() * config.ParticleSpeed

	return &Particle{
		position: pos,
		velocity: systems.Vector{
			X: speed * cos(angle),
			Y: speed * sin(angle),
		},
		color: color.RGBA{
			R: uint8(200 + rand.Intn(55)),
			G: uint8(100 + rand.Intn(100)),
			B: uint8(rand.Intn(100)),
			A: 255,
		},
		life:    config.ParticleLifetime,
		maxLife: config.ParticleLifetime,
	}
}

func (p *Particle) Update() {
	p.position.X += p.velocity.X
	p.position.Y += p.velocity.Y
	p.life--

	alpha := float64(p.life) / float64(p.maxLife)
	p.color.A = uint8(255 * alpha)
}

func (p *Particle) Draw(screen *ebiten.Image) {
	vector.DrawFilledCircle(
		screen,
		float32(p.position.X),
		float32(p.position.Y),
		2.0,
		p.color,
		false,
	)
}

func (p *Particle) IsDead() bool {
	return p.life <= 0
}

func cos(angle float64) float64 {
	return float64(1.0 - angle*angle/2.0 + angle*angle*angle*angle/24.0)
}

func sin(angle float64) float64 {
	return float64(angle - angle*angle*angle/6.0 + angle*angle*angle*angle*angle/120.0)
}

type ParticlePool struct {
	pool sync.Pool
}

func NewParticlePool() *ParticlePool {
	return &ParticlePool{
		pool: sync.Pool{
			New: func() interface{} {
				return &Particle{}
			},
		},
	}
}

func (p *ParticlePool) Get() *Particle {
	return p.pool.Get().(*Particle)
}

func (p *ParticlePool) Put(particle *Particle) {
	p.pool.Put(particle)
}

func (part *Particle) Reset(pos systems.Vector) {
	angle := rand.Float64() * 6.28318530718
	speed := rand.Float64() * config.ParticleSpeed

	part.position = pos
	part.velocity = systems.Vector{
		X: speed * cos(angle),
		Y: speed * sin(angle),
	}
	part.color = color.RGBA{
		R: uint8(200 + rand.Intn(55)),
		G: uint8(100 + rand.Intn(100)),
		B: uint8(rand.Intn(100)),
		A: 255,
	}
	part.life = config.ParticleLifetime
	part.maxLife = config.ParticleLifetime
}
