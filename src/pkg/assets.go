package assets

import (
	"embed"
	"image"
	_ "image/png"
	"io/fs"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

//go:embed *
var assets embed.FS

var PlayerSprite = mustLoadImage("profile/player.png")
var LaserSprite = mustLoadImage("profile/laser.png")
var LaserBeamSprite = mustLoadImage("profile/laserbeam_shot.png")
var GopherPlayer = mustLoadImage("profile/go_player.png")
var PauseIcon = mustLoadImage("profile/pause_icon.png")
var BossSprite = mustLoadImage("boss/boss_ship.png")
var BossTankSprite = mustLoadImage("boss/boss_tank.png")
var BossSniperSprite = mustLoadImage("boss/boss_sniper.png")
var BossSwarmSprite = mustLoadImage("boss/boss_swarm.png")

var SkinGray = mustLoadImage("profile/player.png")
var SkinGreen = mustLoadImage("skins/green.png")
var SkinYellow = mustLoadImage("skins/yellow.png")
var SkinPink = mustLoadImage("skins/pink.png")
var SkinRed = mustLoadImage("skins/red.png")
var SkinPurple = mustLoadImage("skins/purple.png")
var SkinBlack = mustLoadImage("skins/black.png")
var SkinGold = mustLoadImage("skins/gold.png")
var SkinWhite = mustLoadImage("skins/white.png")

var SkinMap = map[string]*ebiten.Image{
	"gray":   SkinGray,
	"green":  SkinGreen,
	"yellow": SkinYellow,
	"pink":   SkinPink,
	"red":    SkinRed,
	"purple": SkinPurple,
	"black":  SkinBlack,
	"gold":   SkinGold,
	"white":  SkinWhite,
}

var MeteorSprites = mustLoadImages("meteors/*.png")
var PlanetsSprites = mustLoadImages("planets/*.png")

var ScoreFont = mustLoadFont("font/font.ttf", 72)
var FontUi = mustLoadFont("font/fontui.ttf", 48)
var FontBtn = mustLoadFont("font/fontbtn.ttf", 48)
var FontSmall = mustLoadFont("font/fontui.ttf", 28)

var PowerUpSprites = mustLoadImage("powers/powerup.png")
var SuperPowerSprite = mustLoadImage("powers/superpower.png")
var HeartPowerUpSprite = mustLoadImage("powers/heart.png")
var HeartUISprite = mustLoadImage("powers/heart.png")
var ExtraLifeUISprite = mustLoadImage("powers/extralife.png")
var ShieldPowerUpSprite = mustLoadImage("powers/shield.png")
var ClockPowerUpSprite = mustLoadImage("powers/clock.png")
var LaserPowerUpSprite = mustLoadImage("powers/laser.png")
var NukePowerUpSprite = mustLoadImage("powers/nuke.png")
var ExtraLifePowerUpSprite = mustLoadImage("powers/extralife.png")
var MultiplierPowerUpSprite = mustLoadImage("powers/multiplier.png")
var CoinSprite = mustLoadImage("profile/coin.png")

var ScrollArrow = mustLoadImage("mobile_controls/scroll_arrow.png")

func mustLoadImage(name string) *ebiten.Image {
	f, err := assets.Open(name)
	if err != nil {
		log.Printf("Error loading image %s: %v", name, err)
		panic(err)
	}

	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		log.Printf("Error decoding image %s: %v", name, err)
		panic(err)
	}

	return ebiten.NewImageFromImage(img)
}

func mustLoadImages(path string) []*ebiten.Image {
	matches, err := fs.Glob(assets, path)
	if err != nil {
		log.Printf("Error loading images from %s: %v", path, err)
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
		log.Printf("Error loading font %s: %v", name, err)
		panic(err)
	}

	tt, err := opentype.Parse(f)
	if err != nil {
		log.Printf("Error parsing font %s: %v", name, err)
		panic(err)
	}

	face, err := opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Printf("Error creating font face for %s: %v", name, err)
		panic(err)
	}

	return face
}
