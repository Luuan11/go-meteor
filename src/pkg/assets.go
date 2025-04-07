package assets

import (
	"bytes"
	"embed"
	"fmt"
	"image"
	_ "image/png"
	"io"
	"io/fs"
	"log"

	"github.com/hajimehoshi/ebiten/audio"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

//go:embed *
var assets embed.FS

var PlayerSprite = mustLoadImage("profile/player.png")
var LaserSprite = mustLoadImage("profile/laser.png")
var GopherPlayer = mustLoadImage("profile/go_player.png")

var MeteorSprites = mustLoadImages("meteors/*.png")
var StarsSprites = mustLoadImages("stars/*.png")
var PlanetsSprites = mustLoadImages("planets/*.png")

var ScoreFont = mustLoadFont("font/font.ttf")
var FontUi = mustLoadFont("font/fontui.ttf")
var FontBtn = mustLoadFont("font/fontbtn.ttf")

var PowerUpSprites = mustLoadImage("powers/powerup.png")
var SuperPowerSprite = mustLoadImage("powers/superpower.png")

var AudioContext *audio.Context
var backgroundMusic *audio.Player

func init() {
	var err error
	AudioContext, err = audio.NewContext(44100)
	if err != nil {
		log.Fatalf("Error creating audio context: %v", err)
	}

	backgroundMusic = mustLoadSound("sounds/music.mp3")
	backgroundMusic.SetVolume(1)
	backgroundMusic.Play()
}

func mustLoadSound(name string) *audio.Player {
	if AudioContext == nil {
		log.Fatal("Audio context not initialized")
	}

	f, err := assets.Open(name)
	if err != nil {
		log.Fatalf("Error opening the file: %v", err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		log.Fatalf("Error reading audio file: %v", err)
	}

	stream, err := mp3.DecodeWithSampleRate(44100, bytes.NewReader(data))
	if err != nil {
		log.Fatalf("Error decoding audio: %v", err)
	}

	player, err := audio.NewPlayer(AudioContext, io.NopCloser(stream))
	if err != nil {
		log.Fatalf("Error creating audio player: %v", err)
	}

	return player
}

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

func mustLoadFont(name string) font.Face {
	f, err := assets.ReadFile(name)
	if err != nil {
		fmt.Println("Error loading font", err)
		panic(err)
	}

	tt, err := opentype.Parse(f)
	if err != nil {
		panic(err)
	}

	face, err := opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    48,
		DPI:     72,
		Hinting: font.HintingVertical,
	})
	if err != nil {
		fmt.Println("Error loading font", err)
		panic(err)
	}

	return face
}
