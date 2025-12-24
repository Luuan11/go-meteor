package core

import (
	"math/rand"
	"time"

	"go-meteor/internal/config"
	"go-meteor/internal/entities"
	"go-meteor/internal/systems"
	"go-meteor/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

func (g *Game) Update() error {
	g.updateStars()

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
	return err
}

func (g *Game) updateMenu() error {
	g.menu.SetScores(g.highScore, g.lastScore)

	touchIDs := ebiten.AppendTouchIDs(nil)
	g.detectMobileTouch(touchIDs)

	g.menu.Update()

	if g.menu.ShouldOpenSettings() {
		g.openSettings(config.StateMenu)
		return nil
	}

	if g.menu.ShouldOpenShop() {
		g.openShop(config.StateMenu)
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
		g.openSettings(config.StatePaused)
	case ui.PauseActionShop:
		g.openShop(config.StatePaused)
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
		g.openShop(config.StateGameOver)
		return nil
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if g.handleGameOverMouseClick() {
			return nil
		}
		g.startNewGame()
		return nil
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.startNewGame()
		return nil
	}

	if g.handleGameOverTouch() {
		return nil
	}

	return nil
}

func (g *Game) startNewGame() {
	g.prepareGameReset()
	g.initNewGameSession()
	g.state = config.StatePlaying
}

func (g *Game) openShopFromGameOver() {
	g.openShop(config.StateGameOver)
}

func (g *Game) handleGameOverMouseClick() bool {
	if g.statistics != nil && g.statistics.CheckSettingsClick() {
		g.openSettings(config.StateGameOver)
		return true
	}

	if g.statistics != nil && g.statistics.CheckShopClick() {
		g.openShop(config.StateGameOver)
		return true
	}

	if g.isMobile && g.isShopButtonClicked(ebiten.CursorPosition()) {
		g.openShop(config.StateGameOver)
		return true
	}

	return false
}

func (g *Game) isShopButtonClicked(x, y int) bool {
	const btnX, btnY, btnSize = 10, 10, 35
	return x >= btnX && x <= btnX+btnSize && y >= btnY && y <= btnY+btnSize
}

func (g *Game) handleGameOverTouch() bool {
	touchIDs := inpututil.AppendJustPressedTouchIDs(nil)
	if len(touchIDs) == 0 {
		return false
	}

	for _, id := range touchIDs {
		x, y := ebiten.TouchPosition(id)
		if g.isShopButtonClicked(x, y) {
			g.openShop(config.StateGameOver)
			return true
		}
	}
	g.startNewGame()
	return true
}

func (g *Game) openShop(previousState config.GameState) {
	g.stateBeforePause = previousState
	g.shop.SetProgress(g.progress)
	g.state = config.StateShop
}

func (g *Game) openSettings(previousState config.GameState) {
	g.stateBeforePause = previousState
	g.settingsMenu.Reset()
	g.state = config.StateSettings
}

func (g *Game) updateShop() error {
	g.shop.SetMobile(g.isMobile)
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
	case ui.ShopActionBuySkin:
		skinID := g.shop.GetSkinID()
		skinCost := g.getSkinCost(skinID)
		if skinCost > 0 && g.progress.BuySkin(skinID, skinCost) {
			g.saveProgress()
			g.shop.SetProgress(g.progress)
		}
	case ui.ShopActionEquipSkin:
		skinID := g.shop.GetSkinID()
		if g.progress.EquipSkin(skinID) {
			g.saveProgress()
			g.shop.SetProgress(g.progress)
			if g.player != nil {
				g.player.SetSkin(skinID)
			}
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

func (g *Game) findSkin(skinID string) *ui.SkinItem {
	for i := range g.shop.Skins {
		if g.shop.Skins[i].ID == skinID {
			return &g.shop.Skins[i]
		}
	}
	return nil
}

func (g *Game) getSkinCost(skinID string) int {
	if skin := g.findSkin(skinID); skin != nil {
		return skin.Cost
	}
	return 0
}

func (g *Game) getSkinItem(skinID string) *ui.SkinItem {
	return g.findSkin(skinID)
}

func (g *Game) updateSettings() error {
	g.settingsMenu.Update()

	if g.settingsMenu.IsClosed() {
		if g.stateBeforePause != 0 {
			g.state = g.stateBeforePause
			g.stateBeforePause = 0
		} else {
			g.state = config.StateMenu
		}
	}

	return nil
}

func (g *Game) updatePlayerDeath() error {
	g.playerDeathTimer++

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
	g.updatePowerTimer(&g.superPowerActive, g.superPowerTimer)
	g.updatePowerTimer(&g.slowMotionActive, g.slowMotionTimer)
	g.updatePowerTimer(&g.laserBeamActive, g.laserBeamTimer)
	g.updatePowerTimer(&g.nukeActive, g.nukeTimer)
	g.updatePowerTimer(&g.multiplierActive, g.multiplierTimer)

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

func (g *Game) updatePowerTimer(active *bool, timer *systems.Timer) {
	if !*active {
		return
	}
	timer.Update()
	if timer.IsReady() {
		timer.Reset()
		*active = false
	}
}
