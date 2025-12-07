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
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

func (g *Game) Update() error {
	g.updateStars()

	stateStart := time.Now()
	var err error
	switch g.state {
	case config.StateMenu:
		err = g.updateMenu()
	case config.StatePlaying:
		err = g.updatePlaying()
	case config.StateBossAnnouncement:
		err = g.updateBossAnnouncement()
	case config.StateBossFight:
		err = g.updateBossFight()
	case config.StatePaused:
		err = g.updatePaused()
	case config.StateGameOver:
		err = g.updateGameOver()
	case config.StateShop:
		err = g.updateShop()
	case config.StateSettings:
		err = g.updateSettings()
	case config.StatePlayerDeath:
		err = g.updatePlayerDeath()
	case config.StateWaitingNameInput:
		return nil
	}

	if time.Since(stateStart) > 10*time.Millisecond {
		fmt.Printf("⚠️ STATE UPDATE SLOW (%v): %v\n", g.state, time.Since(stateStart))
	}

	return err
}

func (g *Game) updateMenu() error {
	g.menu.SetScores(g.highScore, g.lastScore)

	touchIDs := ebiten.AppendTouchIDs(nil)
	g.detectMobileTouch(touchIDs)

	g.menu.Update()

	if g.menu.ShouldOpenSettings() {
		g.stateBeforePause = config.StateMenu
		g.settingsMenu.Reset()
		g.state = config.StateSettings
		return nil
	}

	if g.menu.ShouldOpenShop() {
		g.stateBeforePause = config.StateMenu
		g.shop.SetProgress(g.progress)
		g.state = config.StateShop
		return nil
	}

	if g.menu.IsReady() {
		g.initNewGameSession()
		g.state = config.StatePlaying
	}

	return nil
}

func (g *Game) updatePlaying() error {
	if g.shouldPause() {
		g.stateBeforePause = config.StatePlaying
		g.state = config.StatePaused
		return nil
	}

	if g.shouldSpawnBoss() {
		g.spawnBoss()
		return nil
	}

	touchIDs := ebiten.AppendTouchIDs(nil)
	g.detectMobileTouch(touchIDs)
	g.handleMobileControls(touchIDs)

	g.player.Update()
	g.notification.Update()

	speedMultiplier := 1.0
	if g.wave >= 20 {
		speedMultiplier = 1.0 + float64(g.wave-20)*config.WaveDifficultyFactor
	}
	meteorsPerWave := max(config.MeteorsPerWaveOffset, config.MeteorsPerWaveOffset+(g.wave-1)/config.WaveMeteoIncrement)

	if !g.nukeActive {
		g.updateAndSpawn(g.meteoSpawnTimer, func() {
			for i := 0; i < meteorsPerWave; i++ {
				m := g.meteorPool.Get()
				m.Reset(speedMultiplier)
				g.meteors = append(g.meteors, m)
			}
		})
	}

	g.updateAndSpawn(g.powerUpSpawnTimer, func() {
		var p *entities.PowerUp

		if g.wave >= 5 && rand.Float64() < 0.50 {
			p = entities.NewPowerUpWithType(entities.PowerUpExtraLife)
		} else if g.wave >= config.MinWaveForLaser && rand.Float64() < 0.60 {
			powerType := entities.PowerUpLaser
			roll := rand.Float64()
			if roll < 0.25 {
				powerType = entities.PowerUpNuke
			} else if roll < 0.50 {
				powerType = entities.PowerUpMultiplier
			}
			p = entities.NewPowerUpWithType(powerType)
		} else {
			powerType := entities.PowerUpSuperShot
			roll := rand.Float64()
			if roll < 0.25 {
				powerType = entities.PowerUpShield
			} else if roll < 0.50 {
				powerType = entities.PowerUpSlowMotion
			} else if roll < 0.75 {
				powerType = entities.PowerUpMultiplier
			}
			p = entities.NewPowerUpWithType(powerType)
		}

		g.powerUps = append(g.powerUps, p)
	})

	g.updateGameTimers()

	for _, p := range g.powerUps {
		p.Update()
	}

	g.updateMeteors()
	g.updateAllEntities()
	g.updateBossEntities()

	playerDied := g.checkCollisions()

	if playerDied {
		return nil
	}

	g.player.UpdateTimers()
	g.cleanObjects()

	g.screenShake = max(0, g.screenShake-1)

	return nil
}

func (g *Game) updatePaused() error {
	currentTime := time.Since(g.gameStartTime)
	g.pauseMenu.SetStats(g.score, g.wave, g.meteorsDestroyed, g.powerUpsCollected, currentTime)

	action := g.pauseMenu.Update()

	switch action {
	case ui.PauseActionContinue:
		if g.stateBeforePause != 0 {
			g.state = g.stateBeforePause
		} else {
			g.state = config.StatePlaying
		}
	case ui.PauseActionRestart:
		g.Reset()
		g.state = config.StatePlaying
	case ui.PauseActionQuit:
		g.returnToMenu()
	case ui.PauseActionSettings:
		g.stateBeforePause = config.StatePaused
		g.settingsMenu.Reset()
		g.state = config.StateSettings
	case ui.PauseActionShop:
		g.stateBeforePause = config.StatePaused
		g.shop.SetProgress(g.progress)
		g.state = config.StateShop
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyP) {
		if g.stateBeforePause != 0 {
			g.state = g.stateBeforePause
		} else {
			g.state = config.StatePlaying
		}
	}

	return nil
}

func (g *Game) updateGameOver() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyS) {
		g.stateBeforePause = config.StateGameOver
		g.shop.SetProgress(g.progress)
		g.state = config.StateShop
		return nil
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if g.statistics != nil && g.statistics.CheckSettingsClick() {
			g.settingsMenu.Reset()
			g.state = config.StateSettings
			return nil
		}

		if g.isMobile {
			x, y := ebiten.CursorPosition()
			shopButtonX := (config.ScreenWidth - 180) / 2
			shopButtonY := 535

			if x >= shopButtonX && x <= shopButtonX+180 && y >= shopButtonY && y <= shopButtonY+40 {
				g.shop.SetProgress(g.progress)
				g.state = config.StateShop
				return nil
			}
		}

		g.returnToMenu()
		return nil
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) ||
		inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.returnToMenu()
		return nil
	}

	touchIDs := inpututil.AppendJustPressedTouchIDs(nil)
	if len(touchIDs) > 0 {
		for _, id := range touchIDs {
			x, y := ebiten.TouchPosition(id)
			shopButtonX := (config.ScreenWidth - 180) / 2
			shopButtonY := 535

			if x >= shopButtonX && x <= shopButtonX+180 && y >= shopButtonY && y <= shopButtonY+40 {
				g.stateBeforePause = config.StateGameOver
				g.shop.SetProgress(g.progress)
				g.state = config.StateShop
				return nil
			}
		}
		g.returnToMenu()
	}

	return nil
}

func (g *Game) updateShop() error {
	action := g.shop.Update()

	switch action {
	case ui.ShopActionClose:
		if g.stateBeforePause == config.StatePaused {
			g.state = config.StatePaused
		} else if g.stateBeforePause == config.StateMenu {
			g.state = config.StateMenu
		} else {
			g.state = config.StateGameOver
		}
		g.stateBeforePause = 0
	case ui.ShopActionUpgrade:
		powerType := g.shop.GetUpgradeType()
		item := g.getShopItem(powerType)
		if item != nil && g.progress.Coins >= item.NextCost {
			g.progress.SpendCoins(item.NextCost)
			g.progress.UpgradePower(powerType)
			g.saveProgress()
			g.shop.SetProgress(g.progress)
		}
	}

	return nil
}

func (g *Game) getShopItem(powerType string) *ui.ShopItem {
	for i := range g.shop.Items {
		if g.shop.Items[i].PowerType == powerType {
			return &g.shop.Items[i]
		}
	}
	return nil
}

func (g *Game) updateSettings() error {
	g.settingsMenu.Update()

	if g.settingsMenu.IsClosed() {
		if g.stateBeforePause != 0 {
			g.state = g.stateBeforePause
			g.stateBeforePause = 0
		} else {
			g.state = config.StateMenu
			assets.ResumeMusic()
		}
	}

	return nil
}

func (g *Game) updatePlayerDeath() error {
	g.playerDeathTimer++

	// Create explosion particles continuously
	if g.playerDeathTimer%5 == 0 {
		for i := 0; i < 3; i++ {
			g.createExplosion(
				systems.Vector{
					X: g.playerDeathExplosionX,
					Y: g.playerDeathExplosionY,
				},
				5,
			)
		}
	}

	// After animation duration, proceed to game over
	if g.playerDeathTimer >= config.PlayerDeathAnimationDuration {
		g.survivalTime = time.Since(g.gameStartTime)
		g.statistics = ui.NewStatistics(g.meteorsDestroyed, g.powerUpsCollected, g.wave, g.score, g.survivalTime)
		g.saveHighScore()

		if g.leaderboard.IsTopScore(g.score) && g.hasNameInputModal() {
			g.state = config.StateWaitingNameInput
			g.showNameInputModal()
		} else {
			g.state = config.StateGameOver
		}
	}

	return nil
}

func (g *Game) updateStars() {
	g.updateAndSpawn(g.starSpawnTimer, func() {
		g.stars = append(g.stars, entities.NewStar())
	})

	for _, s := range g.stars {
		s.Update()
	}

	g.cleanStars()
}

func (g *Game) updateMeteors() {
	if g.slowMotionActive {
		for _, m := range g.meteors {
			oldY := m.GetMovement().Y
			m.SetMovementY(oldY * config.SlowMotionFactor)
			m.Update()
			m.SetMovementY(oldY)
		}
	} else {
		for _, m := range g.meteors {
			m.Update()
		}
	}
}

func (g *Game) updateGameTimers() {
	if g.superPowerActive {
		g.superPowerTimer.Update()
		if g.superPowerTimer.IsReady() {
			g.superPowerTimer.Reset()
			g.superPowerActive = false
		}
	}

	if g.slowMotionActive {
		g.slowMotionTimer.Update()
		if g.slowMotionTimer.IsReady() {
			g.slowMotionTimer.Reset()
			g.slowMotionActive = false
		}
	}

	if g.laserBeamActive {
		g.laserBeamTimer.Update()
		if g.laserBeamTimer.IsReady() {
			g.laserBeamTimer.Reset()
			g.laserBeamActive = false
		}
	}

	if g.nukeActive {
		g.nukeTimer.Update()
		if g.nukeTimer.IsReady() {
			g.nukeTimer.Reset()
			g.nukeActive = false
		}
	}

	if g.multiplierActive {
		g.multiplierTimer.Update()
		if g.multiplierTimer.IsReady() {
			g.multiplierTimer.Reset()
			g.multiplierActive = false
		}
	}

	g.comboTimer.Update()
	if g.comboTimer.IsReady() && g.combo > 0 && !g.nukeActive {
		g.combo = 0
	}

	if g.bossDefeated {
		g.bossCooldownTimer.Update()
		if g.bossCooldownTimer.IsReady() {
			g.bossDefeated = false
		}
	}

	if g.isPostBossInvincible {
		g.postBossInvincibilityTimer.Update()
		if g.postBossInvincibilityTimer.IsReady() {
			g.isPostBossInvincible = false
		}
	}
}
