package assets

import (
	"bytes"
	"io"
	"log"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
)

const sampleRate = 48000

var (
	audioContext *audio.Context

	shootSound     []byte
	explosionSound []byte
	powerupSound   []byte
	coinSound      []byte
	damageSound    []byte
	gameoverSound  []byte

	masterVolume = 0.7
	sfxVolume    = 0.7
	sfxEnabled   = true
)

func InitAudio() {
	audioContext = audio.NewContext(sampleRate)

	shootSound = loadSoundBytes("sounds/shoot.wav")
	explosionSound = loadSoundBytes("sounds/explosion.wav")
	powerupSound = loadSoundBytes("sounds/powerup.wav")
	coinSound = loadSoundBytes("sounds/coin.wav")
	damageSound = loadSoundBytes("sounds/damage.wav")
	gameoverSound = loadSoundBytes("sounds/gameover.wav")

	log.Println("Audio system initialized")
}

func loadSoundBytes(path string) []byte {
	soundFile, err := assets.Open(path)
	if err != nil {
		log.Printf("Warning: Could not load sound %s: %v", path, err)
		return nil
	}
	defer soundFile.Close()

	decoded, err := wav.DecodeWithSampleRate(sampleRate, soundFile)
	if err != nil {
		log.Printf("Warning: Could not decode sound %s: %v", path, err)
		return nil
	}

	data, err := io.ReadAll(decoded)
	if err != nil {
		log.Printf("Warning: Could not read sound %s: %v", path, err)
		return nil
	}

	return data
}

func playSound(soundData []byte, volume float64) {
	if !sfxEnabled || soundData == nil || audioContext == nil {
		return
	}

	reader := bytes.NewReader(soundData)
	player, err := audioContext.NewPlayer(reader)
	if err != nil {
		return
	}

	player.SetVolume(volume * sfxVolume * masterVolume)
	player.Play()
}

func PlayShootSound() {
	playSound(shootSound, 0.2)
}

func PlayExplosionSound() {
	playSound(explosionSound, 0.2)
}

func PlayPowerUpSound() {
	playSound(powerupSound, 0.9)
}

func PlayCoinSound() {
	playSound(coinSound, 0.6)
}

func PlayDamageSound() {
	playSound(damageSound, 0.6)
}

func PlayGameOverSound() {
	playSound(gameoverSound, 0.8)
}

func SetMasterVolume(vol float64) {
	masterVolume = clamp(vol, 0, 1)
}

func SetSFXVolume(vol float64) {
	sfxVolume = clamp(vol, 0, 1)
}

func ToggleSFX() {
	sfxEnabled = !sfxEnabled
}

func IsSFXEnabled() bool {
	return sfxEnabled
}

func GetMasterVolume() float64 {
	return masterVolume
}

func GetSFXVolume() float64 {
	return sfxVolume
}

func clamp(val, min, max float64) float64 {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}
