package assets

import (
	"embed"
	"fmt"
	"image"
	_ "image/png"
	"io/fs"

	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

//go:embed *
var assets embed.FS

var PlayerSprite = mustLoadImage("profile/player.png")
var LaserSprite = mustLoadImage("profile/laser.png")
var GopherPlayer = mustLoadImage("profile/go_player.png")
var PauseIcon = mustLoadImage("profile/pause_icon.png")
var BossSprite = mustLoadImage("boss/boss_ship.png")

var MeteorSprites = mustLoadImages("meteors/*.png")
var PlanetsSprites = mustLoadImages("planets/*.png")

var ScoreFont = mustLoadFont("font/font.ttf", 72)
var FontUi = mustLoadFont("font/fontui.ttf", 48)
var FontBtn = mustLoadFont("font/fontbtn.ttf", 48)
var FontSmall = mustLoadFont("font/fontui.ttf", 28)

var PowerUpSprites = mustLoadImage("powers/powerup.png")
var SuperPowerSprite = mustLoadImage("powers/superpower.png")
var HeartPowerUpSprite = mustLoadImage("powers/heart.png")
var ShieldPowerUpSprite = mustLoadImage("powers/shield.png")
var ClockPowerUpSprite = mustLoadImage("powers/clock.png")

func PlayShootSound()     {}
func PlayExplosionSound() {}
func PlayPowerUpSound()   {}
func PlayDamageSound()    {}
func PlayGameOverSound()  {}

func mustLoadImage(name string) *ebiten.Image {
	f, err := assets.Open(name)
	if err != nil {
		fmt.Println("Error loading image", err)
		panic(err)
	}

	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		fmt.Println("Error loading image", err)
		panic(err)
	}

	return ebiten.NewImageFromImage(img)
}

func mustLoadImages(path string) []*ebiten.Image {
	matches, err := fs.Glob(assets, path)
	if err != nil {
		fmt.Println("Error loading images", err)
		panic(err)
	}

	images := make([]*ebiten.Image, len(matches))
	for i, match := range matches {
		images[i] = mustLoadImage(match)
	}

	return images
}

func mustLoadFont(name string, size float64) font.Face {
	f, err := assets.ReadFile(name)
	if err != nil {
		fmt.Println("Error loading font", err)
		panic(err)
	}

	tt, err := opentype.Parse(f)
	if err != nil {
		fmt.Println("Error parsing font", err)
		panic(err)
	}

	face, err := opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		fmt.Println("Error creating font face", err)
		panic(err)
	}

	return face
}
