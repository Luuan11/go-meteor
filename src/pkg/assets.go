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
var PauseIcon = mustLoadImage("profile/pause_icon.png")

var MeteorSprites = mustLoadImages("meteors/*.png")
var StarsSprites = mustLoadImages("stars/*.png")
var PlanetsSprites = mustLoadImages("planets/*.png")

var ScoreFont = mustLoadFont("font/font.ttf", 72)
var FontUi = mustLoadFont("font/fontui.ttf", 48)
var FontBtn = mustLoadFont("font/fontbtn.ttf", 48)
var FontSmall = mustLoadFont("font/fontui.ttf", 28)

var PowerUpSprites = mustLoadImage("powers/powerup.png")
var SuperPowerSprite = mustLoadImage("powers/superpower.png")
var HeartPowerUpSprite = mustLoadImage("powers/heart.png")
var ShieldPowerUpSprite = mustLoadImage("powers/shield.png")

var AudioContext *audio.Context
var backgroundMusic *audio.Player

var shootSoundData []byte
var explosionSoundData []byte
var powerUpSoundData []byte
var damageSoundData []byte
var gameOverSoundData []byte

func init() {
	var err error
	AudioContext, err = audio.NewContext(44100)
	if err != nil {
		log.Fatalf("Error creating audio context: %v", err)
	}

	backgroundMusic = mustLoadSound("sounds/music.mp3")
	if backgroundMusic != nil {
		backgroundMusic.SetVolume(0.3)
		backgroundMusic.Play()
	}

	shootSoundData = tryLoadSoundData("sounds/shoot.mp3")
	explosionSoundData = tryLoadSoundData("sounds/explosion.mp3")
	powerUpSoundData = tryLoadSoundData("sounds/powerup.mp3")
	damageSoundData = tryLoadSoundData("sounds/damage.mp3")
	gameOverSoundData = tryLoadSoundData("sounds/gameover.mp3")
}

func tryLoadSoundData(name string) []byte {
	f, err := assets.Open(name)
	if err != nil {
		return nil
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil
	}

	return data
}

type nopCloser struct {
	io.ReadSeeker
}

func (n nopCloser) Close() error {
	return nil
}

func PlaySound(soundData []byte) {
	if soundData == nil || AudioContext == nil {
		return
	}

	reader := bytes.NewReader(soundData)
	stream, err := mp3.DecodeWithSampleRate(44100, reader)
	if err != nil {
		return
	}

	player, err := audio.NewPlayer(AudioContext, nopCloser{stream})
	if err != nil {
		return
	}

	player.SetVolume(0.8)
	player.Play()
}

func PlayShootSound() {
	PlaySound(shootSoundData)
}

func PlayExplosionSound() {
	PlaySound(explosionSoundData)
}

func PlayPowerUpSound() {
	PlaySound(powerUpSoundData)
}

func PlayDamageSound() {
	PlaySound(damageSoundData)
}

func PlayGameOverSound() {
	PlaySound(gameOverSoundData)
}

func PauseMusic() {
	if backgroundMusic != nil && backgroundMusic.IsPlaying() {
		backgroundMusic.Pause()
	}
}

func ResumeMusic() {
	if backgroundMusic != nil && !backgroundMusic.IsPlaying() {
		backgroundMusic.Play()
	}
}

func SetMusicVolume(volume float64) {
	if backgroundMusic != nil {
		backgroundMusic.SetVolume(volume)
	}
}

func ShouldRestartMusic() bool {
	return backgroundMusic != nil && !backgroundMusic.IsPlaying()
}

func RestartMusic() {
	if backgroundMusic != nil {
		backgroundMusic.Play()
	}
}

func UpdateAudio() {
	if backgroundMusic != nil && !backgroundMusic.IsPlaying() {
		backgroundMusic.Play()
	}
}

func mustLoadSound(name string) *audio.Player {
	f, err := assets.Open(name)
	if err != nil {
		log.Fatalf("Error opening the file %s: %v", name, err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		log.Fatalf("Error reading audio file %s: %v", name, err)
	}

	stream, err := mp3.DecodeWithSampleRate(44100, bytes.NewReader(data))
	if err != nil {
		log.Fatalf("Error decoding audio %s: %v", name, err)
	}

	player, err := audio.NewPlayer(AudioContext, nopCloser{stream})
	if err != nil {
		log.Fatalf("Error creating audio player for %s: %v", name, err)
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
