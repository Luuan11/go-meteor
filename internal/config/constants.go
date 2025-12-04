package config

import "time"

const (
	ScreenWidth  = 800
	ScreenHeight = 600

	PlayerSpeed         = 6.0
	PlayerShootCooldown = time.Millisecond * 500
	PlayerMaxLives      = 3
	PlayerMaxExtraLives = 2
	PlayerMaxTotalLives = 5

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
	SlowMotionTime   = 15 * time.Second
	SlowMotionFactor = 0.25

	LaserBeamTime       = 3500 * time.Millisecond
	NukeClearScreenTime = 5 * time.Second
	MinWaveForLaser     = 5
	MinWaveForNuke      = 5

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

	ScreenShakeDuration   = 10
	ScreenShakeIntensity  = 8.0
	ScreenShakeBossHit    = 5
	ScreenShakeBossDefeat = 20

	BossWaveInterval     = 5
	BossScoreThreshold   = 250
	BossReward           = 100
	BossCooldownTime     = 60 * time.Second
	PowerUpSpawnTimeBoss = 8 * time.Second

	BossTankHealth          = 150
	BossTankSpeed           = 2.0
	BossTankShootCooldown   = time.Millisecond * 1000
	BossTankProjectileSpeed = 4.0

	BossSniperHealth          = 80
	BossSniperSpeed           = 4.0
	BossSniperShootCooldown   = time.Millisecond * 600
	BossSniperProjectileSpeed = 7.0

	BossSwarmHealth          = 100
	BossSwarmSpeed           = 3.5
	BossSwarmShootCooldown   = time.Millisecond * 1200
	BossSwarmProjectileSpeed = 5.0
	BossMinionHealth         = 5
	BossMinionSize           = 15.0
	BossMinionCount          = 3

	// UI Constants
	PauseIconSize   = 30
	PauseIconMargin = 15
	ParticleSize    = 2.0

	// Game Mechanics
	BossTypesCount        = 3
	MeteorsPerWaveOffset  = 1
	WaveMeteoIncrement    = 5
	WaveDifficultyFactor  = 0.15
	BossWarningShakeTime  = 30
	BossScoreProximity    = 10
	ExplosionParticlesMul = 3
	MinionParticles       = 3
)
