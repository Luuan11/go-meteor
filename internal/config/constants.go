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

	// Special Meteor Types
	MeteorIceSlowFactor         = 0.5 // Player moves at 50% speed
	MeteorIceSlowDuration       = 3 * time.Second
	MeteorExplosiveRadius       = 80.0
	MeteorExplosiveDamageRadius = 60.0
	MeteorIceSpawnChance        = 0.08 // 8% chance
	MeteorExplosiveSpawnChance  = 0.06 // 6% chance
	MeteorExplosiveSpeed        = 3.5  // Slower than normal meteors

	LaserSpeed      = 7.0
	SuperLaserSpeed = 12.0

	PowerUpSpeed     = 3.0
	PowerUpSpawnTime = 20 * time.Second
	SuperPowerTime   = 10 * time.Second
	ShieldTime       = 10 * time.Second
	SlowMotionTime   = 15 * time.Second
	SlowMotionFactor = 0.25

	LaserBeamTime       = 3500 * time.Millisecond
	NukeClearScreenTime = 5 * time.Second
	MultiplierTime      = 20 * time.Second
	MultiplierBonus     = 2.0
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

	PlayerDeathAnimationDuration = 90 // 1.5 seconds at 60 FPS
	PlayerDeathExplosionCount    = 30 // Number of explosion particles

	ScreenShakeDuration   = 10
	ScreenShakeIntensity  = 8.0
	ScreenShakeBossHit    = 5
	ScreenShakeBossDefeat = 20

	BossWaveInterval          = 5
	BossScoreThreshold        = 250
	BossReward                = 100
	BossCooldownTime          = 60 * time.Second
	PowerUpSpawnTimeBoss      = 8 * time.Second
	PostBossInvincibilityTime = 3 * time.Second

	BossTankHealth          = 150
	BossTankSpeed           = 2.0
	BossTankShootCooldown   = time.Millisecond * 1000
	BossTankProjectileSpeed = 4.0

	BossSniperHealth          = 80
	BossSniperSpeed           = 4.8
	BossSniperShootCooldown   = time.Millisecond * 600
	BossSniperProjectileSpeed = 7.0

	BossSwarmHealth          = 100
	BossSwarmSpeed           = 5.25
	BossSwarmShootCooldown   = time.Millisecond * 1200
	BossSwarmProjectileSpeed = 5.0
	BossMinionHealth         = 8
	BossMinionSize           = 15.0
	BossMinionCount          = 2
	BossSwarmMinionCount     = 3
	PointsPerMinionKill      = 25

	// UI Constants
	PauseIconSize   = 30
	PauseIconMargin = 15
	ParticleSize    = 2.0

	HeartSpacing         = 40
	HeartOffsetX         = 10
	HeartOffsetY         = 10
	PowerUpBarStartY     = 100
	PowerUpBarSpacing    = 30
	JoystickOffsetX      = 100
	JoystickOffsetY      = 120
	JoystickRadius       = 60
	ShootButtonOffsetX   = 100
	ShootButtonOffsetY   = 120
	ShootButtonRadius    = 50
	BossAnnouncementFade = 30

	// Game Mechanics
	BossTypesCount             = 3
	MeteorsPerWaveOffset       = 1
	WaveMeteoIncrement         = 5
	WaveDifficultyFactor       = 0.15
	BossWarningShakeTime       = 30
	BossScoreProximity         = 10
	BossAnnouncementTime       = 120
	ExplosionParticlesMul      = 3
	MinionParticles            = 3
	BossProjectileOffsetY      = 40
	BossSwarmProjectileOffsetX = 30
	BossShootMinY              = 100
	InitialCapacityLasers      = 100
	InitialCapacityParticles   = 100
)
