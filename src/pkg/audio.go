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

	musicPlayer *audio.Player

	masterVolume = 0.7
	sfxVolume    = 0.7
	musicVolume  = 0.4
	musicEnabled = true
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

	loadBackgroundMusic()

	log.Println("Audio system initialized")
}

func loadBackgroundMusic() {
	musicFile, err := assets.Open("sounds/music.wav")
	if err != nil {
		log.Printf("Warning: Could not load background music: %v", err)
		return
	}

	decoded, err := wav.DecodeWithSampleRate(sampleRate, musicFile)
	if err != nil {
		log.Printf("Warning: Could not decode background music: %v", err)
		return
	}

	loop := audio.NewInfiniteLoop(decoded, decoded.Length())

	musicPlayer, err = audioContext.NewPlayer(loop)
	if err != nil {
		log.Printf("Warning: Could not create music player: %v", err)
		return
	}

	musicPlayer.SetVolume(musicVolume * masterVolume)
	if musicEnabled {
		musicPlayer.Play()
		log.Println("Background music loaded")
	} else {
		log.Println("Background music loaded (but not playing)")
	}
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
	if musicPlayer != nil {
		musicPlayer.SetVolume(musicVolume * masterVolume)
	}
}

func SetSFXVolume(vol float64) {
	sfxVolume = clamp(vol, 0, 1)
}

func SetMusicVolume(vol float64) {
	musicVolume = clamp(vol, 0, 1)
	if musicPlayer != nil {
		musicPlayer.SetVolume(musicVolume * masterVolume)
	}
}

func ToggleSFX() {
	sfxEnabled = !sfxEnabled
}

func ToggleMusic() {
	musicEnabled = !musicEnabled
	if musicPlayer != nil {
		if musicEnabled {
			musicPlayer.Play()
		} else {
			musicPlayer.Pause()
		}
	}
}

func IsSFXEnabled() bool {
	return sfxEnabled
}

func IsMusicEnabled() bool {
	return musicEnabled
}

func GetMasterVolume() float64 {
	return masterVolume
}

func GetSFXVolume() float64 {
	return sfxVolume
}

func GetMusicVolume() float64 {
	return musicVolume
}

func PauseMusic() {
	if musicPlayer != nil {
		musicPlayer.Pause()
	}
}

func ResumeMusic() {
	if musicPlayer != nil && musicEnabled {
		musicPlayer.Play()
	}
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
