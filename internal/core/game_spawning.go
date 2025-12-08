package core

import (
	"time"

	"go-meteor/internal/config"
	"go-meteor/internal/effects"
	"go-meteor/internal/entities"
	"go-meteor/internal/input"
	"go-meteor/internal/systems"
	"go-meteor/internal/ui"
	assets "go-meteor/src/pkg"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Spawning and object management functions

func (g *Game) updateAndSpawn(timer *systems.Timer, spawnFunc func()) {
	timer.Update()
	if timer.IsReady() {
		timer.Reset()
		spawnFunc()
	}
}

func (g *Game) createExplosion(pos systems.Vector, count int) {
	if g.isMobile {
		count = count / 2
		if count < 3 {
			count = 3
		}
	}
	for i := 0; i < count; i++ {
		g.particles = append(g.particles, effects.NewParticle(pos))
	}
}

func (g *Game) handlePowerUpCollected(powerType entities.PowerUpType) {
	g.powerUpsCollected++
	switch powerType {
	case entities.PowerUpSuperShot:
		g.superPowerActive = true
		duration := config.SuperPowerTime + (g.getUpgradeBonus("superpower") * time.Second)
		g.superPowerTimer = systems.NewTimer(duration)
		g.notification.Show("SUPER POWER!", ui.NotificationSuperPower)
	case entities.PowerUpHeart:
		g.player.Heal()
		g.notification.Show("+1 LIFE", ui.NotificationLife)
	case entities.PowerUpShield:
		duration := config.ShieldTime + (g.getUpgradeBonus("shield") * time.Second)
		g.player.ActivateShieldWithDuration(duration)
		g.notification.Show("SHIELD ACTIVE", ui.NotificationShield)
	case entities.PowerUpSlowMotion:
		g.slowMotionActive = true
		duration := config.SlowMotionTime + (g.getUpgradeBonus("slowmotion") * time.Second)
		g.slowMotionTimer = systems.NewTimer(duration)
		g.notification.Show("SLOW MOTION!", ui.NotificationSuperPower)
	case entities.PowerUpLaser:
		g.laserBeamActive = true
		duration := config.LaserBeamTime + (g.getUpgradeBonus("laser") * time.Second)
		g.laserBeamTimer = systems.NewTimer(duration)
		g.notification.Show("LASER BEAM!", ui.NotificationSuperPower)
	case entities.PowerUpNuke:
		g.activateNuke()
		g.notification.Show("NUKE ACTIVATED!", ui.NotificationSuperPower)
	case entities.PowerUpExtraLife:
		g.player.GainExtraLife()
		g.notification.Show("EXTRA LIFE!", ui.NotificationLife)
	case entities.PowerUpMultiplier:
		g.multiplierActive = true
		duration := config.MultiplierTime + (g.getUpgradeBonus("multiplier") * time.Second)
		g.multiplierTimer = systems.NewTimer(duration)
		g.notification.Show("SCORE x2!", ui.NotificationSuperPower)
	}
}

func (g *Game) getUpgradeBonus(powerType string) time.Duration {
	if g.progress == nil {
		return 0
	}
	level := g.progress.GetUpgradeLevel(powerType)
	return time.Duration(level * 2)
}

func (g *Game) activateNuke() {
	meteorsDestroyed := len(g.meteors)

	for _, m := range g.meteors {
		g.createExplosion(m.GetPosition(), 5)
		g.meteorPool.Put(m)
	}

	g.meteors = g.meteors[:0]

	g.addScore(meteorsDestroyed * 2)
	g.addScreenShake(20)
	g.nukeActive = true
	g.nukeTimer.Reset()
	assets.PlayExplosionSound()
}

func (g *Game) cleanObjects() {
	validMeteors := make([]*entities.Meteor, 0, len(g.meteors))
	for _, m := range g.meteors {
		if m.IsOutOfScreen() {
			g.meteorPool.Put(m)
		} else {
			validMeteors = append(validMeteors, m)
		}
	}
	g.meteors = validMeteors

	validLasers := make([]*entities.Laser, 0, len(g.lasers))
	for _, l := range g.lasers {
		if l.IsOutOfScreen() {
			g.laserPool.Put(l)
		} else {
			validLasers = append(validLasers, l)
		}
	}
	g.lasers = validLasers

	validPowerUps := make([]*entities.PowerUp, 0, len(g.powerUps))
	for _, p := range g.powerUps {
		if p.IsOutOfScreen() {
			g.powerUpPool.Put(p)
		} else {
			validPowerUps = append(validPowerUps, p)
		}
	}
	g.powerUps = validPowerUps

	validParticles := make([]*effects.Particle, 0, len(g.particles))
	for _, p := range g.particles {
		if !p.IsDead() {
			validParticles = append(validParticles, p)
		}
	}
	g.particles = validParticles

	validCoins := make([]*entities.Coin, 0, len(g.coins))
	for _, c := range g.coins {
		if !c.IsOffScreen() {
			validCoins = append(validCoins, c)
		}
	}
	g.coins = validCoins
}

func (g *Game) cleanStars() {
	validStars := make([]*entities.Star, 0, len(g.stars))
	for _, s := range g.stars {
		if !s.IsOutOfScreen() {
			validStars = append(validStars, s)
		}
	}
	g.stars = validStars
}

func (g *Game) clearPools() {
	for _, m := range g.meteors {
		g.meteorPool.Put(m)
	}
	for _, l := range g.lasers {
		g.laserPool.Put(l)
	}
	for _, p := range g.powerUps {
		g.powerUpPool.Put(p)
	}
	for _, bp := range g.bossProjectiles {
		g.bossProjectilePool.Put(bp)
	}
}

func (g *Game) returnToMenu() {
	g.clearPools()
	g.lastScore = g.score
	g.meteors = g.meteors[:0]
	g.lasers = g.lasers[:0]
	g.powerUps = g.powerUps[:0]
	g.particles = g.particles[:0]
	g.bossProjectiles = g.bossProjectiles[:0]
	g.player = entities.NewPlayer(g)
	g.Reset()
	g.menu.Reset()
	g.state = config.StateMenu
}

func (g *Game) Reset() {
	g.clearPools()

	g.player = entities.NewPlayer(g)
	g.meteors = g.meteors[:0]
	g.lasers = g.lasers[:0]
	g.powerUps = g.powerUps[:0]
	g.particles = g.particles[:0]
	g.bossProjectiles = g.bossProjectiles[:0]
	g.meteoSpawnTimer.Reset()
	g.starSpawnTimer.Reset()
	g.powerUpSpawnTimer.Reset()
	g.comboTimer.Reset()
	g.bossCooldownTimer.Reset()

	g.saveHighScore()

	g.score = 0
	g.combo = 0
	g.wave = 1
	g.superPowerActive = false
	g.slowMotionActive = false
	g.slowMotionTimer.Reset()
	g.superPowerTimer.Reset()
	g.laserBeamActive = false
	g.laserBeamTimer.Reset()
	g.nukeActive = false
	g.nukeTimer.Reset()
	g.multiplierActive = false
	g.multiplierTimer.Reset()
	g.screenShake = 0
	g.boss = nil
	g.bossDefeated = false
	g.bossWarningShown = false
	g.bossNoDamage = false
	g.bossCount = 0
	g.initNewGameSession()
}

func (g *Game) handleGameOver() bool {
	collider := g.player.Collider()
	g.playerDeathExplosionX = collider.X + collider.Width/2
	g.playerDeathExplosionY = collider.Y + collider.Height/2
	g.playerDeathTimer = 0

	// Create initial large explosion
	g.createExplosion(
		systems.Vector{X: g.playerDeathExplosionX, Y: g.playerDeathExplosionY},
		config.PlayerDeathExplosionCount,
	)
	g.screenShake = config.ScreenShakeBossDefeat

	assets.PlayExplosionSound()
	assets.PlayGameOverSound()

	g.state = config.StatePlayerDeath
	return true
}

func (g *Game) addScore(basePoints int) {
	points := basePoints + int(float64(g.combo)*config.ComboMultiplier)
	if points < 1 {
		points = 1
	}

	if g.multiplierActive {
		points = int(float64(points) * config.MultiplierBonus)
	}

	g.score += points

	if g.score >= g.wave*config.WaveScoreThreshold {
		g.wave++
	}
}

func (g *Game) addScreenShake(intensity int) {
	g.screenShake = intensity
}

func (g *Game) loadHighScore() {
	g.highScore = g.storage.LoadHighScore()
}

func (g *Game) saveHighScore() {
	if g.score > g.highScore {
		g.highScore = g.score
		g.storage.SaveHighScore(g.highScore)
	}
}

func (g *Game) loadLeaderboard() {
	data, err := g.storage.LoadLeaderboard()
	if err == nil {
		g.leaderboard.FromJSON(data)
	}
}

func (g *Game) loadProgress() {
	progress, err := g.storage.LoadProgress()
	if err == nil {
		g.progress = progress
	} else {
		g.progress = systems.NewPlayerProgress()
	}
}

func (g *Game) saveProgress() {
	if g.progress != nil {
		g.storage.SaveProgress(g.progress)
	}
}

func (g *Game) initMobileControls() {
	if g.joystick == nil {
		g.joystick = input.NewJoystick(config.JoystickOffsetX, float64(config.ScreenHeight-config.JoystickOffsetY), config.JoystickRadius)
	}
	if g.shootButton == nil {
		g.shootButton = input.NewShootButton(float64(config.ScreenWidth-config.ShootButtonOffsetX), float64(config.ScreenHeight-config.ShootButtonOffsetY), config.ShootButtonRadius)
	}
}

func (g *Game) shouldPause() bool {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyP) {
		return true
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		if g.isPauseIconClicked(x, y) {
			return true
		}
	}

	touchIDs := ebiten.AppendTouchIDs(nil)
	for _, id := range touchIDs {
		x, y := ebiten.TouchPosition(id)
		if g.isPauseIconClicked(x, y) {
			return true
		}
	}

	return false
}

func (g *Game) isPauseIconClicked(x, y int) bool {
	const iconSize = 30
	return x >= g.pauseIconX && x <= g.pauseIconX+iconSize &&
		y >= g.pauseIconY && y <= g.pauseIconY+iconSize
}

func (g *Game) handleMobileControls(touchIDs []ebiten.TouchID) {
	if !g.isMobile || g.joystick == nil || g.shootButton == nil {
		return
	}

	g.joystick.Update(touchIDs)

	// Shoot continuously while button is pressed (like desktop)
	if g.shootButton.Update(touchIDs) || g.shootButton.IsActive() {
		g.player.Shoot()
	}

	dx, dy := g.joystick.GetDirection()
	if !g.joystick.IsPressed() {
		return
	}

	const threshold = 0.3
	if dx < -threshold {
		g.player.MoveLeft()
	}
	if dx > threshold {
		g.player.MoveRight()
	}
	if dy < -threshold {
		g.player.MoveUp()
	}
	if dy > threshold {
		g.player.MoveDown()
	}
}

// Helper functions

func (g *Game) AddLaser(l *entities.Laser) {
	g.lasers = append(g.lasers, l)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return config.ScreenWidth, config.ScreenHeight
}

func (g *Game) GetSuperPowerActive() bool {
	return g.superPowerActive
}

func (g *Game) GetLaserBeamActive() bool {
	return g.laserBeamActive
}

func (g *Game) ResetCombo() {
	g.combo = 0
	g.comboTimer.Reset()
}
