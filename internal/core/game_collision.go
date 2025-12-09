package core

import (
	"go-meteor/internal/config"
	"go-meteor/internal/entities"
	"go-meteor/internal/systems"
	"go-meteor/internal/ui"
	assets "go-meteor/src/pkg"
)

func (g *Game) checkCollisions() bool {
	g.checkLaserMeteorCollisions()
	return g.checkPlayerCollisions()
}

func (g *Game) checkLaserMeteorCollisions() {
	meteorsToRemove := make(map[int]bool)
	lasersToRemove := make(map[int]bool)

	for i := range g.meteors {
		if meteorsToRemove[i] {
			continue
		}
		for j := range g.lasers {
			if lasersToRemove[j] {
				continue
			}
			if g.meteors[i].Collider().Intersects(g.lasers[j].Collider()) {
				g.handleMeteorDestruction(i, j, meteorsToRemove, lasersToRemove)
			}
		}
	}

	g.filterMeteors(meteorsToRemove)
	g.filterLasers(lasersToRemove)
}

func (g *Game) handleMeteorDestruction(meteorIdx, laserIdx int, meteorsToRemove, lasersToRemove map[int]bool) {
	meteorType := g.meteors[meteorIdx].GetType()
	meteorPos := g.meteors[meteorIdx].GetPosition()

	if meteorType == entities.MeteorExplosive {
		g.handleExplosiveMeteor(meteorPos, meteorsToRemove)
	} else {
		g.createExplosion(meteorPos, config.ParticleCount)
	}

	meteorsToRemove[meteorIdx] = true
	g.meteorsDestroyed++
	if !g.laserBeamActive {
		lasersToRemove[laserIdx] = true
	}

	g.combo++
	g.comboTimer.Reset()
	g.addScore(1)
	assets.PlayExplosionSound()
}

func (g *Game) checkPlayerCollisions() bool {
	if g.checkMeteorPlayerCollision() {
		return true
	}
	g.checkPowerUpPlayerCollision()
	g.checkCoinPlayerCollision()
	return false
}

func (g *Game) checkMeteorPlayerCollision() bool {
	for i := len(g.meteors) - 1; i >= 0; i-- {
		if !g.meteors[i].Collider().Intersects(g.player.Collider()) {
			continue
		}

		if g.isPostBossInvincible {
			continue
		}

		meteorType := g.meteors[i].GetType()
		meteorPos := g.meteors[i].GetPosition()

		if meteorType == entities.MeteorIce {
			g.handleIceMeteorCollision(i, meteorPos)
		} else if meteorType == entities.MeteorExplosive {
			return g.handleExplosiveMeteorCollision(i, meteorPos)
		} else {
			if g.handleNormalMeteorCollision(i, meteorPos) {
				return true
			}
		}
		break
	}
	return false
}

func (g *Game) handleIceMeteorCollision(meteorIdx int, meteorPos systems.Vector) {
	g.player.ApplySlow()
	g.notification.Show("FROZEN!", ui.NotificationShield)
	g.createExplosion(meteorPos, config.ParticleCount)

	g.meteorPool.Put(g.meteors[meteorIdx])
	g.meteors = append(g.meteors[:meteorIdx], g.meteors[meteorIdx+1:]...)
	g.addScreenShake(config.ScreenShakeDuration)
}

func (g *Game) handleExplosiveMeteorCollision(meteorIdx int, meteorPos systems.Vector) bool {
	isDead := g.player.TakeDamage()

	g.meteorPool.Put(g.meteors[meteorIdx])
	g.meteors = append(g.meteors[:meteorIdx], g.meteors[meteorIdx+1:]...)

	g.handleExplosiveMeteorDirect(meteorPos)

	return isDead && g.handleGameOver()
}

func (g *Game) handleNormalMeteorCollision(meteorIdx int, meteorPos systems.Vector) bool {
	isDead := g.player.TakeDamage()
	g.createExplosion(meteorPos, config.ParticleCount)

	if isDead {
		g.meteorPool.Put(g.meteors[meteorIdx])
		g.meteors = append(g.meteors[:meteorIdx], g.meteors[meteorIdx+1:]...)
		return g.handleGameOver()
	}

	g.meteorPool.Put(g.meteors[meteorIdx])
	g.meteors = append(g.meteors[:meteorIdx], g.meteors[meteorIdx+1:]...)
	g.addScreenShake(config.ScreenShakeDuration)
	return false
}

func (g *Game) checkPowerUpPlayerCollision() {
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

func (g *Game) checkCoinPlayerCollision() {
	for i := len(g.coins) - 1; i >= 0; i-- {
		if g.coins[i].IsCollected() {
			if g.coins[i].HasReachedTarget() {
				coinValue := g.coins[i].GetValue()
				g.progress.AddCoins(coinValue)
				g.saveProgress()
				g.coins = append(g.coins[:i], g.coins[i+1:]...)
				assets.PlayCoinSound()
			}
			continue
		}

		coinX, coinY, coinW, coinH := g.coins[i].GetBounds()
		playerCollider := g.player.Collider()

		if coinX < playerCollider.X+playerCollider.Width && coinX+coinW > playerCollider.X &&
			coinY < playerCollider.Y+playerCollider.Height && coinY+coinH > playerCollider.Y {

			g.coins[i].Collect(35, 93)
			break
		}
	}
}

func (g *Game) filterMeteors(toRemove map[int]bool) {
	newMeteors := make([]*entities.Meteor, 0, len(g.meteors))
	for i, m := range g.meteors {
		if toRemove[i] {
			if entities.ShouldDropCoin() {
				coin := entities.NewCoinFromMeteor(m)
				g.coins = append(g.coins, coin)
			}
			g.meteorPool.Put(m)
		} else {
			newMeteors = append(newMeteors, m)
		}
	}
	g.meteors = newMeteors
}

func (g *Game) filterLasers(toRemove map[int]bool) {
	newLasers := make([]*entities.Laser, 0, len(g.lasers))
	for i, l := range g.lasers {
		if toRemove[i] {
			g.laserPool.Put(l)
		} else {
			newLasers = append(newLasers, l)
		}
	}
	g.lasers = newLasers
}

func (g *Game) handleExplosiveMeteor(explosionPos systems.Vector, meteorsToRemove map[int]bool) {
	g.createExplosion(explosionPos, config.ParticleCount*3)
	g.addScreenShake(15)

	for i := range g.meteors {
		if meteorsToRemove[i] {
			continue
		}

		meteorPos := g.meteors[i].GetPosition()
		dx := meteorPos.X - explosionPos.X
		dy := meteorPos.Y - explosionPos.Y
		distance := (dx*dx + dy*dy)

		if distance < config.MeteorExplosiveDamageRadius*config.MeteorExplosiveDamageRadius {
			g.createExplosion(meteorPos, 5)
			g.addScore(1)
			g.meteorsDestroyed++
			meteorsToRemove[i] = true
		}
	}
}

func (g *Game) handleExplosiveMeteorDirect(explosionPos systems.Vector) {

	g.createExplosion(explosionPos, config.ParticleCount*3)
	g.addScreenShake(15)

	for i := len(g.meteors) - 1; i >= 0; i-- {
		meteorPos := g.meteors[i].GetPosition()
		dx := meteorPos.X - explosionPos.X
		dy := meteorPos.Y - explosionPos.Y
		distance := (dx*dx + dy*dy)

		if distance < config.MeteorExplosiveDamageRadius*config.MeteorExplosiveDamageRadius {
			g.createExplosion(meteorPos, 5)
			g.addScore(1)
			g.meteorsDestroyed++

			g.meteorPool.Put(g.meteors[i])
			g.meteors = append(g.meteors[:i], g.meteors[i+1:]...)
		}
	}
}
