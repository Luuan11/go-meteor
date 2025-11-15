package config

import "time"

const (
	ScreenWidth  = 800
	ScreenHeight = 600

	PlayerSpeed         = 6.0
	PlayerShootCooldown = time.Millisecond * 500

	MeteorMinSpeed    = 2.0
	MeteorMaxSpeed    = 13.0
	MeteorSpawnTime   = 1 * time.Second
	MeteorRotationMin = -0.02
	MeteorRotationMax = 0.02

	LaserSpeed      = 7.0
	SuperLaserSpeed = 12.0

	PowerUpSpeed     = 6.0
	PowerUpSpawnTime = 20 * time.Second
	SuperPowerTime   = 10 * time.Second
	ShieldTime       = 10 * time.Second

	StarSpawnTime   = (1 * time.Second) / 2
	PlanetSpawnTime = 5 * time.Second

	InitialLives      = 3
	InvincibilityTime = 2 * time.Second

	ComboTimeout    = 3 * time.Second
	ComboMultiplier = 0.5

	WaveMeteoIncrease  = 2
	WaveScoreThreshold = 50

	ParticleLifetime = 30
	ParticleCount    = 15
	ParticleSpeed    = 3.0

	ScreenShakeDuration  = 10
	ScreenShakeIntensity = 8.0
)
