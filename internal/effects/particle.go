package effects

import (
	"go-meteor/internal/config"
	"go-meteor/internal/systems"
	"image/color"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
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
	size := 3.0
	for x := 0.0; x < size; x++ {
		for y := 0.0; y < size; y++ {
			screen.Set(int(p.position.X+x), int(p.position.Y+y), p.color)
		}
	}
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
