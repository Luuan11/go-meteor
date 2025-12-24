package core

import (
	"fmt"
	"math/rand"
	"time"

	"go-meteor/internal/config"
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

	for _, m := range g.meteors {
		m.Update()
	}

	for _, l := range g.lasers {
		l.Update()
	}

	for _, p := range g.particles {
		p.Update()
	}

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

	g.updateBossAndMinions()
	g.spawnBossPowerUps()
	g.updateBossFightObjects()
	g.checkBossCollisions()
	g.cleanBossObjects()

	g.player.UpdateTimers()
	g.screenShake = max(0, g.screenShake-1)

	return nil
}

func (g *Game) updateBossAndMinions() {
	if g.boss == nil {
		return
	}

	playerPos := g.player.Collider()
	g.boss.SetPlayerPosition(systems.Vector{X: playerPos.X, Y: playerPos.Y})
	g.boss.Update()

	g.updateMinions(playerPos)
	g.handleBossShooting()
}

func (g *Game) updateMinions(playerPos systems.Rect) {
	for _, minion := range g.boss.GetMinions() {
		if minion == nil {
			continue
		}

		minion.SetTarget(systems.Vector{X: playerPos.X, Y: playerPos.Y})
		minion.Update()

		if minion.CanShoot() {
			minion.Shoot()
			minionPos := minion.GetPosition()
			bp := g.bossProjectilePool.Get()
			const projectileYOffset = 10.0
			bp.Reset(minionPos.X, minionPos.Y+projectileYOffset)
			g.bossProjectiles = append(g.bossProjectiles, bp)
		}
	}
}

func (g *Game) handleBossShooting() {
	if !g.boss.CanShoot() || g.boss.GetPosition().Y < config.BossShootMinY {
		return
	}

	g.boss.Shoot()
	pos := g.boss.GetPosition()

	if g.boss.GetBossType() == config.BossSwarm {
		g.spawnSwarmProjectiles(pos)
	} else {
		g.spawnSingleProjectile(pos)
	}

	assets.PlayExplosionSound()
}

func (g *Game) spawnSwarmProjectiles(pos systems.Vector) {
	bp1 := g.bossProjectilePool.Get()
	bp1.Reset(pos.X-config.BossSwarmProjectileOffsetX, pos.Y+config.BossProjectileOffsetY)
	g.bossProjectiles = append(g.bossProjectiles, bp1)

	bp2 := g.bossProjectilePool.Get()
	bp2.Reset(pos.X+config.BossSwarmProjectileOffsetX, pos.Y+config.BossProjectileOffsetY)
	g.bossProjectiles = append(g.bossProjectiles, bp2)
}

func (g *Game) spawnSingleProjectile(pos systems.Vector) {
	bp := g.bossProjectilePool.Get()
	bp.Reset(pos.X, pos.Y+config.BossProjectileOffsetY)
	g.bossProjectiles = append(g.bossProjectiles, bp)
}

func (g *Game) spawnBossPowerUps() {
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
}

func (g *Game) updateBossFightObjects() {
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
}

func (g *Game) checkBossCollisions() {
	if g.boss == nil {
		return
	}

	g.checkLaserBossCollisions()
	g.checkPowerUpCollisionsBoss()
	g.checkBossProjectileCollisions()
	g.checkMinionPlayerCollision()
}

func (g *Game) checkLaserBossCollisions() {
	if g.boss == nil {
		return
	}
	for i := len(g.lasers) - 1; i >= 0; i-- {
		if i >= len(g.lasers) {
			continue
		}
		if g.checkLaserHitBoss(i) {
			return
		}
		g.checkLaserHitMinions(i)
	}
}

func (g *Game) checkLaserHitBoss(laserIdx int) bool {
	if !g.lasers[laserIdx].Collider().Intersects(g.boss.Collider()) {
		return false
	}

	damage := g.lasers[laserIdx].GetDamage()
	isDead := g.boss.TakeDamage(damage)

	if !g.lasers[laserIdx].IsLaserBeam() {
		g.createExplosion(g.boss.GetPosition(), 5)
		g.laserPool.Put(g.lasers[laserIdx])
		g.lasers = append(g.lasers[:laserIdx], g.lasers[laserIdx+1:]...)
	}

	g.addScreenShake(config.ScreenShakeBossHit)
	assets.PlayExplosionSound()

	if isDead {
		g.defeatBoss()
		return true
	}
	return false
}

func (g *Game) checkLaserHitMinions(laserIdx int) {
	if g.boss == nil || laserIdx >= len(g.lasers) {
		return
	}
	minions := g.boss.GetMinions()
	if minions == nil {
		return
	}
	for mIdx, minion := range minions {
		if minion == nil || laserIdx >= len(g.lasers) || !g.lasers[laserIdx].Collider().Intersects(minion.Collider()) {
			continue
		}

		damage := g.lasers[laserIdx].GetDamage()
		isDead := minion.TakeDamage(damage)

		if !g.lasers[laserIdx].IsLaserBeam() {
			g.createExplosion(minion.GetPosition(), config.MinionParticles)
			g.laserPool.Put(g.lasers[laserIdx])
			g.lasers = append(g.lasers[:laserIdx], g.lasers[laserIdx+1:]...)
		}

		assets.PlayExplosionSound()

		if isDead {
			g.score += config.PointsPerMinionKill
			g.boss.RemoveMinion(mIdx)
		}
		break
	}
}

func (g *Game) checkPowerUpCollisionsBoss() {
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
}

func (g *Game) checkBossProjectileCollisions() {
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
}

func (g *Game) checkMinionPlayerCollision() {
	if g.boss == nil {
		return
	}
	minions := g.boss.GetMinions()
	if minions == nil {
		return
	}
	for mIdx, minion := range minions {
		if minion == nil || !minion.Collider().Intersects(g.player.Collider()) {
			continue
		}

		isDead := g.player.TakeDamage()
		g.bossNoDamage = false
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

	g.cleanPowerUps()
	g.cleanLasers()
	g.cleanParticles()
}
