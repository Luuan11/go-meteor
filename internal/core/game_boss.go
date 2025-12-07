package core

import (
	"fmt"
	"math/rand"
	"time"

	"go-meteor/internal/config"
	"go-meteor/internal/effects"
	"go-meteor/internal/entities"
	"go-meteor/internal/systems"
	"go-meteor/internal/ui"
	assets "go-meteor/src/pkg"

	"github.com/hajimehoshi/ebiten/v2"
)

// Boss management functions

func (g *Game) shouldSpawnBoss() bool {
	if g.boss != nil || g.bossDefeated {
		return false
	}

	return (g.wave > 0 && g.wave%config.BossWaveInterval == 0 && !g.bossWarningShown) ||
		(g.score >= config.BossScoreThreshold && g.score < config.BossScoreThreshold+config.BossScoreProximity && !g.bossWarningShown)
}

func (g *Game) spawnBoss() {
	g.bossWarningShown = true
	g.screenShake = config.BossWarningShakeTime
	g.bossAnnouncementTimer = config.BossAnnouncementTime
	g.state = config.StateBossAnnouncement
	assets.PlayExplosionSound()
}

func (g *Game) updateBossAnnouncement() error {
	g.bossAnnouncementTimer--

	for _, s := range g.stars {
		s.Update()
	}
	g.cleanStars()

	// Atualiza meteoros durante anúncio para não travarem
	for _, m := range g.meteors {
		m.Update()
	}

	for _, l := range g.lasers {
		l.Update()
	}

	for _, p := range g.particles {
		p.Update()
	}

	// Atualiza power-ups e coins para não travarem
	for _, pu := range g.powerUps {
		pu.Update()
	}

	for _, c := range g.coins {
		c.Update()
	}

	g.cleanObjects()

	g.player.Update()
	g.notification.Update()

	if g.bossAnnouncementTimer <= 0 {
		bossType := config.BossType(rand.Intn(config.BossTypesCount))
		g.boss = entities.NewBoss(bossType)
		g.bossNoDamage = true
		g.bossBar.Show()
		g.state = config.StateBossFight
		g.powerUpSpawnTimer = systems.NewTimer(config.PowerUpSpawnTimeBoss)
	}

	return nil
}

func (g *Game) updateBossFight() error {
	if g.shouldPause() {
		g.stateBeforePause = config.StateBossFight
		g.state = config.StatePaused
		return nil
	}

	touchIDs := ebiten.AppendTouchIDs(nil)
	g.handleMobileControls(touchIDs)

	g.player.Update()
	g.notification.Update()

	g.updateGameTimers()

	if g.boss != nil {
		playerPos := g.player.Collider()
		g.boss.SetPlayerPosition(systems.Vector{X: playerPos.X, Y: playerPos.Y})

		if g.slowMotionActive {
			g.boss.Update()
		} else {
			g.boss.Update()
		}

		// Atualiza e faz minions atirarem
		for _, minion := range g.boss.GetMinions() {
			if minion != nil {
				minion.SetTarget(systems.Vector{X: playerPos.X, Y: playerPos.Y})
				minion.Update()

				// Minion atira
				if minion.CanShoot() {
					minion.Shoot()
					minionPos := minion.GetPosition()
					bp := g.bossProjectilePool.Get()
					bp.Reset(minionPos.X, minionPos.Y+10)
					g.bossProjectiles = append(g.bossProjectiles, bp)
				}
			}
		}

		if g.boss.CanShoot() && g.boss.GetPosition().Y >= config.BossShootMinY {
			g.boss.Shoot()
			pos := g.boss.GetPosition()

			if g.boss.GetBossType() == config.BossSwarm {
				bp1 := g.bossProjectilePool.Get()
				bp1.Reset(pos.X-config.BossSwarmProjectileOffsetX, pos.Y+config.BossProjectileOffsetY)
				g.bossProjectiles = append(g.bossProjectiles, bp1)

				bp2 := g.bossProjectilePool.Get()
				bp2.Reset(pos.X+config.BossSwarmProjectileOffsetX, pos.Y+config.BossProjectileOffsetY)
				g.bossProjectiles = append(g.bossProjectiles, bp2)
			} else {
				bp := g.bossProjectilePool.Get()
				bp.Reset(pos.X, pos.Y+config.BossProjectileOffsetY)
				g.bossProjectiles = append(g.bossProjectiles, bp)
			}

			assets.PlayExplosionSound()
		}
	}

	g.updateAndSpawn(g.powerUpSpawnTimer, func() {
		var p *entities.PowerUp

		if g.wave >= config.MinWaveForLaser && rand.Float64() < 0.60 {
			powerType := entities.PowerUpLaser
			if rand.Float64() < 0.5 {
				powerType = entities.PowerUpNuke
			}
			p = entities.NewPowerUpWithType(powerType)
		} else {
			p = g.powerUpPool.Get()
			p.Reset()
		}

		g.powerUps = append(g.powerUps, p)
	})

	for _, pu := range g.powerUps {
		pu.Update()
	}

	for _, bp := range g.bossProjectiles {
		bp.Update()
	}

	for _, l := range g.lasers {
		l.Update()
	}

	for _, p := range g.particles {
		p.Update()
	}

	g.checkBossCollisions()
	g.cleanBossObjects()

	g.player.UpdateTimers()
	g.screenShake = max(0, g.screenShake-1)

	return nil
}

func (g *Game) checkBossCollisions() {
	if g.boss == nil {
		return
	}

	// Laser-Boss collisions
	for i := len(g.lasers) - 1; i >= 0; i-- {
		if g.lasers[i].Collider().Intersects(g.boss.Collider()) {
			damage := g.lasers[i].GetDamage()
			isDead := g.boss.TakeDamage(damage)

			g.createExplosion(g.boss.GetPosition(), 5)

			if !g.lasers[i].IsLaserBeam() {
				g.laserPool.Put(g.lasers[i])
				g.lasers = append(g.lasers[:i], g.lasers[i+1:]...)
			}

			g.addScreenShake(config.ScreenShakeBossHit)
			assets.PlayExplosionSound()

			if isDead {
				g.defeatBoss()
				return
			}
			continue
		}

		// Laser-Minion collisions (todos os bosses agora têm minions)
		for mIdx, minion := range g.boss.GetMinions() {
			if minion != nil && g.lasers[i].Collider().Intersects(minion.Collider()) {
				damage := g.lasers[i].GetDamage()
				isDead := minion.TakeDamage(damage)

				g.createExplosion(minion.GetPosition(), config.MinionParticles)

				if !g.lasers[i].IsLaserBeam() {
					g.laserPool.Put(g.lasers[i])
					g.lasers = append(g.lasers[:i], g.lasers[i+1:]...)
				}

				assets.PlayExplosionSound()

				if isDead {
					g.score += config.PointsPerMinionKill
					g.boss.RemoveMinion(mIdx)
				}
				break
			}
		}
	}

	// PowerUp-Player collisions
	for i := len(g.powerUps) - 1; i >= 0; i-- {
		if g.powerUps[i].Collider().Intersects(g.player.Collider()) {
			powerType := g.powerUps[i].GetType()
			g.powerUpPool.Put(g.powerUps[i])
			g.powerUps = append(g.powerUps[:i], g.powerUps[i+1:]...)

			g.handlePowerUpCollected(powerType)
			assets.PlayPowerUpSound()
			break
		}
	}

	for i := len(g.bossProjectiles) - 1; i >= 0; i-- {
		if g.bossProjectiles[i].Collider().Intersects(g.player.Collider()) {
			isDead := g.player.TakeDamage()
			g.bossNoDamage = false

			g.bossProjectilePool.Put(g.bossProjectiles[i])
			g.bossProjectiles = append(g.bossProjectiles[:i], g.bossProjectiles[i+1:]...)

			g.addScreenShake(config.ScreenShakeDuration)

			if isDead {
				g.handleGameOver()
				return
			}
			break
		}
	}

	// Minion-Player collision
	for mIdx, minion := range g.boss.GetMinions() {
		if minion != nil && minion.Collider().Intersects(g.player.Collider()) {
			isDead := g.player.TakeDamage()
			g.bossNoDamage = false

			// Minion também morre ao colidir
			g.createExplosion(minion.GetPosition(), config.MinionParticles)
			g.boss.RemoveMinion(mIdx)
			assets.PlayExplosionSound()

			g.addScreenShake(config.ScreenShakeDuration)

			if isDead {
				g.handleGameOver()
				return
			}
			break
		}
	}
}

func (g *Game) defeatBoss() {
	g.createExplosion(g.boss.GetPosition(), config.ParticleCount*config.ExplosionParticlesMul)

	baseReward := config.BossReward

	fightDuration := time.Since(g.boss.GetSpawnTime()).Seconds()
	if fightDuration < 30 {
		timeBonus := int((30 - fightDuration) * 2)
		baseReward += timeBonus
		g.notification.Show(fmt.Sprintf("+%d TIME BONUS!", timeBonus), ui.NotificationLife)
	}

	g.score += baseReward
	g.addScreenShake(config.ScreenShakeBossDefeat)
	g.notification.Show(fmt.Sprintf("+%d BOSS DEFEATED!", baseReward), ui.NotificationSuperPower)
	assets.PlayExplosionSound()

	numPowerUps := 1
	if g.bossNoDamage {
		numPowerUps = 2
		g.notification.Show("PERFECT! +EXTRA POWER-UP", ui.NotificationSuperPower)
	}

	for i := 0; i < numPowerUps; i++ {
		p := g.powerUpPool.Get()
		p.Reset()
		g.powerUps = append(g.powerUps, p)
	}

	for _, bp := range g.bossProjectiles {
		g.bossProjectilePool.Put(bp)
	}

	g.boss = nil
	g.bossProjectiles = nil
	g.bossBar.Hide()
	g.bossWarningShown = false
	g.bossDefeated = true
	g.bossCount++
	g.bossCooldownTimer.Reset()
	g.powerUpSpawnTimer = systems.NewTimer(config.PowerUpSpawnTime)

	// Activate post-boss invincibility
	g.isPostBossInvincible = true
	g.postBossInvincibilityTimer.Reset()
	g.notification.Show("INVINCIBLE!", ui.NotificationShield)

	g.state = config.StatePlaying
}

func (g *Game) cleanBossObjects() {
	validBossProjectiles := make([]*entities.BossProjectile, 0, len(g.bossProjectiles))
	for _, bp := range g.bossProjectiles {
		if bp.IsOutOfScreen() {
			g.bossProjectilePool.Put(bp)
		} else {
			validBossProjectiles = append(validBossProjectiles, bp)
		}
	}
	g.bossProjectiles = validBossProjectiles

	validPowerUps := make([]*entities.PowerUp, 0, len(g.powerUps))
	for _, pu := range g.powerUps {
		if pu.IsOutOfScreen() {
			g.powerUpPool.Put(pu)
		} else {
			validPowerUps = append(validPowerUps, pu)
		}
	}
	g.powerUps = validPowerUps

	validLasers := make([]*entities.Laser, 0, len(g.lasers))
	for _, l := range g.lasers {
		if l.IsOutOfScreen() {
			g.laserPool.Put(l)
		} else {
			validLasers = append(validLasers, l)
		}
	}
	g.lasers = validLasers

	validParticles := make([]*effects.Particle, 0, len(g.particles))
	for _, p := range g.particles {
		if !p.IsDead() {
			validParticles = append(validParticles, p)
		}
	}
	g.particles = validParticles
}
