package entities

import (
	"go-meteor/internal/config"
	"go-meteor/internal/systems"
	"image/color"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var starColor = color.RGBA{255, 255, 100, 255}

var starSprites [3]*ebiten.Image
var starSpritesInitialized = false

type Star struct {
	position   systems.Vector
	movement   systems.Vector
	size       float32
	spriteType int 
}

func initStarSprites() {
	if starSpritesInitialized {
		return
	}

	sizes := []float32{1.0, 1.75, 2.5}
	for i, size := range sizes {
		radius := int(size) + 2
		img := ebiten.NewImage(radius*2, radius*2)
		vector.DrawFilledCircle(img, float32(radius), float32(radius), size, starColor, false)
		starSprites[i] = img
	}
	starSpritesInitialized = true
}

func NewStar() *Star {
	initStarSprites()

	pos := systems.Vector{
		X: rand.Float64() * config.ScreenWidth,
		Y: -10,
	}

	velocity := 2.0 + rand.Float64()*2.0

	movement := systems.Vector{
		X: 0,
		Y: velocity,
	}

	size := 1.0 + rand.Float32()*1.5

	spriteType := 0
	if size > 2.0 {
		spriteType = 2 
	} else if size > 1.5 {
		spriteType = 1 
	}

	m := &Star{
		position:   pos,
		movement:   movement,
		size:       size,
		spriteType: spriteType,
	}
	return m
}

func (m *Star) IsOutOfScreen() bool {
	return m.position.Y > config.ScreenHeight+100
}

func (m *Star) Update() {
	m.position.X += m.movement.X
	m.position.Y += m.movement.Y
}

func (m *Star) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	sprite := starSprites[m.spriteType]
	bounds := sprite.Bounds()
	op.GeoM.Translate(
		m.position.X-float64(bounds.Dx())/2,
		m.position.Y-float64(bounds.Dy())/2,
	)
	screen.DrawImage(sprite, op)
}
