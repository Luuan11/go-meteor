package core

import (
	"fmt"
	"image/color"

	"go-meteor/internal/config"
	"go-meteor/internal/effects"
	"go-meteor/internal/entities"
	"go-meteor/internal/input"
	"go-meteor/internal/systems"
	"go-meteor/internal/ui"
	assets "go-meteor/src/pkg"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
)

type Game struct {
	state config.GameState

	meteoSpawnTimer   *systems.Timer
	starSpawnTimer    *systems.Timer
	powerUpSpawnTimer *systems.Timer
	superPowerTimer   *systems.Timer
	comboTimer        *systems.Timer
	menu              *ui.Menu
	pauseMenu         *ui.PauseMenu
	notification      *ui.Notification

	player           *entities.Player
	meteors          []*entities.Meteor
	stars            []*entities.Star
	lasers           []*entities.Laser
	powerUps         []*entities.PowerUp
	particles        []*effects.Particle
	superPowerActive bool

	boss              *entities.Boss
	bossProjectiles   []*entities.BossProjectile
	bossBar           *ui.BossBar
	bossWarningShown  bool
	bossCooldownTimer *systems.Timer
	bossDefeated      bool

	meteorPool         *entities.MeteorPool
	laserPool          *entities.LaserPool
	powerUpPool        *entities.PowerUpPool
	bossProjectilePool *entities.BossProjectilePool
	particlePool       *effects.ParticlePool

	score int
	combo int
	wave  int

	screenShake int

	joystick      *input.Joystick
	shootButton   *input.ShootButton
	isMobile      bool
	touchDetected bool

	pauseIconX int
	pauseIconY int

	leaderboard *systems.Leaderboard
	storage     systems.Storage
	highScore   int
	lastScore   int
}

func NewGame() *Game {
	g := &Game{
		state:              config.StateMenu,
		meteoSpawnTimer:    systems.NewTimer(config.MeteorSpawnTime),
		starSpawnTimer:     systems.NewTimer(config.StarSpawnTime),
		powerUpSpawnTimer:  systems.NewTimer(config.PowerUpSpawnTime),
		superPowerTimer:    systems.NewTimer(config.SuperPowerTime),
		comboTimer:         systems.NewTimer(config.ComboTimeout),
		bossCooldownTimer:  systems.NewTimer(config.BossCooldownTime),
		superPowerActive:   false,
		meteorPool:         entities.NewMeteorPool(),
		laserPool:          entities.NewLaserPool(),
		powerUpPool:        entities.NewPowerUpPool(),
		bossProjectilePool: entities.NewBossProjectilePool(),
		particlePool:       effects.NewParticlePool(),
		notification:       ui.NewNotification(),
		wave:               1,
		isMobile:           false,
		touchDetected:      false,
		leaderboard:        systems.NewLeaderboard(),
		storage:            systems.NewStorage(),
		meteors:            make([]*entities.Meteor, 0, 50),
		stars:              make([]*entities.Star, 0, 50),
		lasers:             make([]*entities.Laser, 0, 100),
		powerUps:           make([]*entities.PowerUp, 0, 10),
		particles:          make([]*effects.Particle, 0, 100),
		bossProjectiles:    make([]*entities.BossProjectile, 0, 20),
		bossBar:            ui.NewBossBar(),
		bossWarningShown:   false,
		pauseIconX:         config.ScreenWidth - 45,
		pauseIconY:         15,
	}

	g.player = entities.NewPlayer(g)
	g.menu = ui.NewMenu()
	g.pauseMenu = ui.NewPauseMenu()

	g.loadHighScore()
	g.loadLeaderboard()

	return g
}

func (g *Game) Update() error {
	if assets.ShouldRestartMusic() {
		assets.RestartMusic()
	}

	g.updateStars()

	switch g.state {
	case config.StateMenu:
		return g.updateMenu()
	case config.StatePlaying:
		return g.updatePlaying()
	case config.StateBossFight:
		return g.updateBossFight()
	case config.StatePaused:
		return g.updatePaused()
	case config.StateGameOver:
		return g.updateGameOver()
	case config.StateWaitingNameInput:
		return nil
	}

	return nil
}

func (g *Game) updateMenu() error {
	g.menu.SetScores(g.highScore, g.lastScore)

	var touchIDs []ebiten.TouchID
	touchIDs = ebiten.AppendTouchIDs(touchIDs)

	if len(touchIDs) > 0 && !g.touchDetected {
		g.touchDetected = true
		g.isMobile = true
		g.joystick = input.NewJoystick(100, float64(config.ScreenHeight-120), 60)
		g.shootButton = input.NewShootButton(float64(config.ScreenWidth-100), float64(config.ScreenHeight-120), 50)
	}

	g.menu.Update()

	if g.menu.IsReady() {
		g.initNewGameSession()
		g.state = config.StatePlaying
	}

	return nil
}

func (g *Game) updatePlaying() error {
	if g.shouldPause() {
		assets.PauseMusic()
		g.state = config.StatePaused
		return nil
	}

	if g.shouldSpawnBoss() {
		g.spawnBoss()
		return nil
	}

	touchIDs := ebiten.AppendTouchIDs(nil)

	if len(touchIDs) > 0 && !g.touchDetected {
		g.touchDetected = true
		g.isMobile = true
	}

	g.handleMobileControls(touchIDs)
	g.player.Update()
	g.notification.Update()

	speedMultiplier := 1.0 + float64(g.wave-1)*0.15
	meteorsPerWave := max(1, 1+(g.wave-1)/5)

	// Limita meteoros para evitar acúmulo excessivo
	if len(g.meteors) < 50 {
		g.updateAndSpawn(g.meteoSpawnTimer, func() {
			for i := 0; i < meteorsPerWave; i++ {
				m := g.meteorPool.Get()
				m.Reset(speedMultiplier)
				g.meteors = append(g.meteors, m)
			}
		})
	}

	g.updateAndSpawn(g.powerUpSpawnTimer, func() {
		p := g.powerUpPool.Get()
		p.Reset()
		g.powerUps = append(g.powerUps, p)
	})

	if g.superPowerActive {
		g.superPowerTimer.Update()
		if g.superPowerTimer.IsReady() {
			g.superPowerTimer.Reset()
			g.superPowerActive = false
		}
	}

	g.comboTimer.Update()
	if g.comboTimer.IsReady() && g.combo > 0 {
		g.combo = 0
	}

	if g.bossDefeated {
		g.bossCooldownTimer.Update()
		if g.bossCooldownTimer.IsReady() {
			g.bossDefeated = false
		}
	}

	for _, p := range g.powerUps {
		p.Update()
	}

	for _, m := range g.meteors {
		m.Update()
	}

	for _, b := range g.lasers {
		b.Update()
	}

	for _, p := range g.particles {
		p.Update()
	}

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
	action := g.pauseMenu.Update()

	switch action {
	case ui.PauseActionContinue:
		assets.ResumeMusic()
		g.state = config.StatePlaying
	case ui.PauseActionRestart:
		g.Reset()
		assets.ResumeMusic()
		g.state = config.StatePlaying
	case ui.PauseActionQuit:
		g.returnToMenu()
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyP) {
		assets.ResumeMusic()
		g.state = config.StatePlaying
	}

	return nil
}

func (g *Game) updateGameOver() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) ||
		inpututil.IsKeyJustPressed(ebiten.KeySpace) ||
		inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		g.returnToMenu()
		return nil
	}

	touchIDs := inpututil.AppendJustPressedTouchIDs(nil)
	if len(touchIDs) > 0 {
		g.returnToMenu()
	}

	return nil
}

func (g *Game) updateStars() {
	// Limita o número máximo de estrelas para evitar acúmulo
	if len(g.stars) < 30 {
		g.updateAndSpawn(g.starSpawnTimer, func() {
			g.stars = append(g.stars, entities.NewStar())
		})
	}

	for _, s := range g.stars {
		s.Update()
	}

	g.cleanStars()
}

func (g *Game) checkCollisions() bool {
	// Otimização: marca para remoção ao invés de remover durante iteração
	meteorsToRemove := make([]int, 0, 10)
	lasersToRemove := make([]int, 0, 10)

	for i := range g.meteors {
		if i >= len(g.meteors) {
			continue
		}

		for j := range g.lasers {
			if j >= len(g.lasers) {
				continue
			}

			if g.meteors[i].Collider().Intersects(g.lasers[j].Collider()) {
				pos := g.meteors[i].GetPosition()

				// Limita partículas para evitar acúmulo
				if len(g.particles) < 80 {
					for k := 0; k < config.ParticleCount; k++ {
						p := g.particlePool.Get()
						p.Reset(pos)
						g.particles = append(g.particles, p)
					}
				}

				meteorsToRemove = append(meteorsToRemove, i)
				lasersToRemove = append(lasersToRemove, j)

				g.combo++
				g.comboTimer.Reset()
				points := 1 + int(float64(g.combo)*config.ComboMultiplier)
				if points < 0 {
					points = 1
				}
				g.score += points

				if g.score < 0 {
					g.score = 0
				}

				assets.PlayExplosionSound()

				if g.score >= g.wave*config.WaveScoreThreshold {
					g.wave++
				}

				break
			}
		}
	}

	// Remove lasers em batch
	if len(lasersToRemove) > 0 {
		lasersAlive := 0
		removeMap := make(map[int]bool)
		for _, idx := range lasersToRemove {
			removeMap[idx] = true
		}
		for i := 0; i < len(g.lasers); i++ {
			if !removeMap[i] {
				g.lasers[lasersAlive] = g.lasers[i]
				lasersAlive++
			} else {
				g.laserPool.Put(g.lasers[i])
			}
		}
		g.lasers = g.lasers[:lasersAlive]
	}

	// Remove meteors em batch
	if len(meteorsToRemove) > 0 {
		meteorsAlive := 0
		removeMap := make(map[int]bool)
		for _, idx := range meteorsToRemove {
			removeMap[idx] = true
		}
		for i := 0; i < len(g.meteors); i++ {
			if !removeMap[i] {
				g.meteors[meteorsAlive] = g.meteors[i]
				meteorsAlive++
			} else {
				g.meteorPool.Put(g.meteors[i])
			}
		}
		g.meteors = g.meteors[:meteorsAlive]
	}

	// Colisão com jogador
	for i := len(g.meteors) - 1; i >= 0; i-- {
		if i >= len(g.meteors) {
			continue
		}

		if g.meteors[i].Collider().Intersects(g.player.Collider()) {
			isDead := g.player.TakeDamage()

			g.meteorPool.Put(g.meteors[i])
			// Remove mantendo ordem
			meteorsAlive := 0
			for j := 0; j < len(g.meteors); j++ {
				if j != i {
					g.meteors[meteorsAlive] = g.meteors[j]
					meteorsAlive++
				}
			}
			g.meteors = g.meteors[:meteorsAlive]

			if isDead {
				g.saveHighScore()
				if g.leaderboard.IsTopScore(g.score) && g.hasNameInputModal() {
					g.state = config.StateWaitingNameInput
					g.showNameInputModal()
				} else {
					g.state = config.StateGameOver
				}
				return true
			}
			g.screenShake = config.ScreenShakeDuration
			break
		}
	}

	for i := len(g.powerUps) - 1; i >= 0; i-- {
		if i >= len(g.powerUps) {
			continue
		}

		if g.powerUps[i].Collider().Intersects(g.player.Collider()) {
			powerType := g.powerUps[i].GetType()
			g.powerUpPool.Put(g.powerUps[i])
			g.powerUps = append(g.powerUps[:i], g.powerUps[i+1:]...)

			switch powerType {
			case entities.PowerUpSuperShot:
				g.superPowerActive = true
				g.superPowerTimer.Reset()
				g.notification.Show("SUPER POWER!", ui.NotificationSuperPower)
			case entities.PowerUpHeart:
				g.player.Heal()
				g.notification.Show("+1 LIFE", ui.NotificationLife)
			case entities.PowerUpShield:
				g.player.ActivateShield()
				g.notification.Show("SHIELD ACTIVE", ui.NotificationShield)
			}

			assets.PlayPowerUpSound()
			break
		}
	}

	return false
}

func (g *Game) cleanObjects() {
	// Otimização: remove sem realocação usando swap com último elemento
	meteorsAlive := 0
	for i := 0; i < len(g.meteors); i++ {
		if !g.meteors[i].IsOutOfScreen() {
			g.meteors[meteorsAlive] = g.meteors[i]
			meteorsAlive++
		} else {
			g.meteorPool.Put(g.meteors[i])
		}
	}
	g.meteors = g.meteors[:meteorsAlive]

	lasersAlive := 0
	for i := 0; i < len(g.lasers); i++ {
		if !g.lasers[i].IsOutOfScreen() {
			g.lasers[lasersAlive] = g.lasers[i]
			lasersAlive++
		} else {
			g.laserPool.Put(g.lasers[i])
		}
	}
	g.lasers = g.lasers[:lasersAlive]

	powerUpsAlive := 0
	for i := 0; i < len(g.powerUps); i++ {
		if !g.powerUps[i].IsOutOfScreen() {
			g.powerUps[powerUpsAlive] = g.powerUps[i]
			powerUpsAlive++
		} else {
			g.powerUpPool.Put(g.powerUps[i])
		}
	}
	g.powerUps = g.powerUps[:powerUpsAlive]

	particlesAlive := 0
	for i := 0; i < len(g.particles); i++ {
		if !g.particles[i].IsDead() {
			g.particles[particlesAlive] = g.particles[i]
			particlesAlive++
		} else {
			g.particlePool.Put(g.particles[i])
		}
	}
	g.particles = g.particles[:particlesAlive]
}

func (g *Game) cleanStars() {
	starsAlive := 0
	for i := 0; i < len(g.stars); i++ {
		if !g.stars[i].IsOutOfScreen() {
			g.stars[starsAlive] = g.stars[i]
			starsAlive++
		}
	}
	g.stars = g.stars[:starsAlive]
}

func (g *Game) Draw(screen *ebiten.Image) {
	offsetX := 0.0
	offsetY := 0.0

	if g.screenShake > 0 {
		offsetX = (float64(g.screenShake%2)*2 - 1) * config.ScreenShakeIntensity
		offsetY = (float64((g.screenShake+1)%2)*2 - 1) * config.ScreenShakeIntensity
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(offsetX, offsetY)

	for _, s := range g.stars {
		s.Draw(screen)
	}

	switch g.state {
	case config.StateMenu:
		g.drawMenu(screen)
	case config.StatePlaying:
		g.drawPlaying(screen)
	case config.StateBossFight:
		g.drawBossFight(screen)
	case config.StatePaused:
		g.drawPlaying(screen)
		g.pauseMenu.Draw(screen)
	case config.StateGameOver:
		g.drawGameOver(screen)
	case config.StateWaitingNameInput:
		g.drawPlaying(screen)
	}
}

func (g *Game) drawMenu(screen *ebiten.Image) {
	g.menu.Draw(screen)
}

func (g *Game) drawPlaying(screen *ebiten.Image) {
	g.player.Draw(screen)

	for _, m := range g.meteors {
		m.Draw(screen)
	}

	for _, b := range g.lasers {
		b.Draw(screen)
	}

	for _, p := range g.powerUps {
		p.Draw(screen)
	}

	for _, p := range g.particles {
		p.Draw(screen)
	}

	g.drawUI(screen)
	g.notification.Draw(screen)
}

func (g *Game) drawUI(screen *ebiten.Image) {
	for i := 0; i < config.InitialLives; i++ {
		filled := i < g.player.GetLives()
		ui.DrawHeart(screen, 20+i*25, 20, filled)
	}

	waveText := fmt.Sprintf("Wave: %d", g.wave)
	drawText(screen, waveText, assets.FontSmall, 20, 65, color.White)

	if g.combo > 1 {
		comboText := fmt.Sprintf("%d COMBO", g.combo)
		comboColor := color.RGBA{255, 200, 0, 255}
		drawText(screen, comboText, assets.FontSmall, 20, 105, comboColor)
	}

	scoreText := fmt.Sprintf("Points: %d", g.score)
	drawText(screen, scoreText, assets.FontUi, 20, 570, color.White)

	highScoreText := fmt.Sprintf("HIGH SCORE: %d", g.highScore)
	highScoreWidth := measureText(highScoreText, assets.FontUi)
	highScoreX := config.ScreenWidth - highScoreWidth - 20
	drawText(screen, highScoreText, assets.FontUi, highScoreX, 570, color.White)

	ui.DrawPauseIcon(screen, g.pauseIconX, g.pauseIconY)

	barY := float32(100)

	if g.superPowerActive {
		progress := float32(g.superPowerTimer.Progress())
		ui.DrawPowerUpBar(screen, "SUPER POWER", progress, color.RGBA{255, 100, 255, 255})
		barY += 30
	}

	if g.player.HasShield() {
		progress := float32(g.player.ShieldProgress())
		ui.DrawPowerUpBarAt(screen, progress, color.RGBA{100, 200, 255, 255}, barY)
	}

	if g.isMobile && g.joystick != nil && g.shootButton != nil {
		g.joystick.Draw(screen)
		g.shootButton.Draw(screen)
	}
}

func (g *Game) drawBossFight(screen *ebiten.Image) {
	g.player.Draw(screen)

	if g.boss != nil {
		g.boss.Draw(screen)
	}

	for _, bp := range g.bossProjectiles {
		bp.Draw(screen)
	}

	for _, b := range g.lasers {
		b.Draw(screen)
	}

	for _, p := range g.particles {
		p.Draw(screen)
	}

	for i := 0; i < config.InitialLives; i++ {
		filled := i < g.player.GetLives()
		ui.DrawHeart(screen, 20+i*25, 20, filled)
	}

	if g.boss != nil {
		g.bossBar.Draw(screen, g.boss.GetHealth(), g.boss.GetMaxHealth())
	}

	g.notification.Draw(screen)

	ui.DrawPauseIcon(screen, g.pauseIconX, g.pauseIconY)

	scoreText := fmt.Sprintf("Points: %d", g.score)
	drawText(screen, scoreText, assets.FontUi, 20, 570, color.White)

	if g.isMobile && g.joystick != nil && g.shootButton != nil {
		g.joystick.Draw(screen)
		g.shootButton.Draw(screen)
	}
}

func (g *Game) drawGameOver(screen *ebiten.Image) {
	youDiedText := "YOU DIED"
	youDiedWidth := measureText(youDiedText, assets.FontUi)
	youDiedX := (config.ScreenWidth - youDiedWidth) / 2
	drawText(screen, youDiedText, assets.FontUi, youDiedX, 300, color.White)

	tryAgainText := "Press ENTER to try again"
	tryAgainWidth := measureText(tryAgainText, assets.FontUi)
	tryAgainX := (config.ScreenWidth - tryAgainWidth) / 2
	drawText(screen, tryAgainText, assets.FontUi, tryAgainX, 400, color.White)

	scoreText := fmt.Sprintf("Points: %d", g.score)
	drawText(screen, scoreText, assets.FontUi, 20, 570, color.White)

	highScoreText := fmt.Sprintf("HIGH SCORE: %d", g.highScore)
	highScoreWidth := measureText(highScoreText, assets.FontUi)
	highScoreX := config.ScreenWidth - highScoreWidth - 20
	drawText(screen, highScoreText, assets.FontUi, highScoreX, 570, color.White)
}

func (g *Game) AddLaser(l *entities.Laser) {
	g.lasers = append(g.lasers, l)
}

func (g *Game) updateAndSpawn(timer *systems.Timer, spawnFunc func()) {
	timer.Update()
	if timer.IsReady() {
		timer.Reset()
		spawnFunc()
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
	if g.shootButton.Update(touchIDs) {
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
	for _, p := range g.particles {
		g.particlePool.Put(p)
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

	g.saveHighScore()

	g.score = 0
	g.combo = 0
	g.wave = 1
	g.superPowerActive = false
	g.screenShake = 0
	g.initNewGameSession()
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return config.ScreenWidth, config.ScreenHeight
}

func (g *Game) GetSuperPowerActive() bool {
	return g.superPowerActive
}

func (g *Game) ResetCombo() {
	g.combo = 0
	g.comboTimer.Reset()
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

func drawText(screen *ebiten.Image, txt string, face font.Face, x, y int, clr color.Color) {
	text.Draw(screen, txt, face, x, y, clr)
}

func measureText(txt string, face font.Face) int {
	bounds := text.BoundString(face, txt)
	return bounds.Dx()
}

func (g *Game) shouldSpawnBoss() bool {
	if g.boss != nil || g.bossDefeated {
		return false
	}

	return (g.wave > 0 && g.wave%config.BossWaveInterval == 0 && !g.bossWarningShown) ||
		(g.score >= config.BossScoreThreshold && g.score < config.BossScoreThreshold+10 && !g.bossWarningShown)
}

func (g *Game) spawnBoss() {
	g.bossWarningShown = true
	g.screenShake = 30
	g.notification.Show("⚠️ BOSS APPROACHING!", ui.NotificationWarning)
	assets.PlayExplosionSound()

	g.boss = entities.NewBoss()
	g.bossBar.Show()
	g.state = config.StateBossFight
}

func (g *Game) updateBossFight() error {
	if g.shouldPause() {
		assets.PauseMusic()
		g.state = config.StatePaused
		return nil
	}

	touchIDs := ebiten.AppendTouchIDs(nil)
	g.handleMobileControls(touchIDs)

	g.player.Update()
	g.notification.Update()

	if g.boss != nil {
		g.boss.Update()

		if g.boss.CanShoot() && g.boss.GetPosition().Y >= 100 {
			g.boss.Shoot()
			pos := g.boss.GetPosition()
			bp := g.bossProjectilePool.Get()
			bp.Reset(pos.X, pos.Y+40)
			g.bossProjectiles = append(g.bossProjectiles, bp)
			assets.PlayExplosionSound()
		}
	}

	for _, bp := range g.bossProjectiles {
		bp.Update()
	}

	for _, b := range g.lasers {
		b.Update()
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

	// Otimização: processa hit e remove depois
	laserHitIndex := -1
	for i := range g.lasers {
		if g.lasers[i].Collider().Intersects(g.boss.Collider()) {
			isDead := g.boss.TakeDamage(1)

			pos := g.boss.GetPosition()
			// Limita partículas
			if len(g.particles) < 80 {
				for k := 0; k < 5; k++ {
					p := g.particlePool.Get()
					p.Reset(pos)
					g.particles = append(g.particles, p)
				}
			}

			laserHitIndex = i
			g.screenShake = 5
			assets.PlayExplosionSound()

			if isDead {
				g.laserPool.Put(g.lasers[i])
				lasersAlive := 0
				for j := 0; j < len(g.lasers); j++ {
					if j != i {
						g.lasers[lasersAlive] = g.lasers[j]
						lasersAlive++
					}
				}
				g.lasers = g.lasers[:lasersAlive]
				g.defeatBoss()
				return
			}
			break
		}
	}

	if laserHitIndex >= 0 {
		g.laserPool.Put(g.lasers[laserHitIndex])
		lasersAlive := 0
		for j := 0; j < len(g.lasers); j++ {
			if j != laserHitIndex {
				g.lasers[lasersAlive] = g.lasers[j]
				lasersAlive++
			}
		}
		g.lasers = g.lasers[:lasersAlive]
	}

	projectileHitIndex := -1
	for i := range g.bossProjectiles {
		if g.bossProjectiles[i].Collider().Intersects(g.player.Collider()) {
			isDead := g.player.TakeDamage()
			projectileHitIndex = i
			g.screenShake = config.ScreenShakeDuration

			if isDead {
				g.bossProjectilePool.Put(g.bossProjectiles[i])
				projectilesAlive := 0
				for j := 0; j < len(g.bossProjectiles); j++ {
					if j != i {
						g.bossProjectiles[projectilesAlive] = g.bossProjectiles[j]
						projectilesAlive++
					}
				}
				g.bossProjectiles = g.bossProjectiles[:projectilesAlive]
				g.saveHighScore()
				if g.leaderboard.IsTopScore(g.score) && g.hasNameInputModal() {
					g.state = config.StateWaitingNameInput
					g.showNameInputModal()
				} else {
					g.state = config.StateGameOver
				}
				return
			}
			break
		}
	}

	if projectileHitIndex >= 0 {
		g.bossProjectilePool.Put(g.bossProjectiles[projectileHitIndex])
		projectilesAlive := 0
		for j := 0; j < len(g.bossProjectiles); j++ {
			if j != projectileHitIndex {
				g.bossProjectiles[projectilesAlive] = g.bossProjectiles[j]
				projectilesAlive++
			}
		}
		g.bossProjectiles = g.bossProjectiles[:projectilesAlive]
	}
}

func (g *Game) defeatBoss() {
	pos := g.boss.GetPosition()
	// Limita partículas
	if len(g.particles) < 80 {
		for k := 0; k < config.ParticleCount*3; k++ {
			p := g.particlePool.Get()
			p.Reset(pos)
			g.particles = append(g.particles, p)
		}
	}

	g.score += config.BossReward
	g.screenShake = 20
	g.notification.Show(fmt.Sprintf("+%d BOSS DEFEATED!", config.BossReward), ui.NotificationSuperPower)
	assets.PlayExplosionSound()

	p := g.powerUpPool.Get()
	p.Reset()
	g.powerUps = append(g.powerUps, p)

	for _, bp := range g.bossProjectiles {
		g.bossProjectilePool.Put(bp)
	}

	g.boss = nil
	g.bossProjectiles = nil
	g.bossBar.Hide()
	g.bossWarningShown = false
	g.bossDefeated = true
	g.bossCooldownTimer.Reset()
	g.state = config.StatePlaying
}

func (g *Game) cleanBossObjects() {
	projectilesAlive := 0
	for i := 0; i < len(g.bossProjectiles); i++ {
		if !g.bossProjectiles[i].IsOutOfScreen() {
			g.bossProjectiles[projectilesAlive] = g.bossProjectiles[i]
			projectilesAlive++
		} else {
			g.bossProjectilePool.Put(g.bossProjectiles[i])
		}
	}
	g.bossProjectiles = g.bossProjectiles[:projectilesAlive]
}
