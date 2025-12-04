package core

import (
	"fmt"
	"image/color"
	"math/rand"
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
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
)

var emptyImage = ebiten.NewImage(1, 1)

func init() {
	emptyImage.Fill(color.White)
}

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
	slowMotionActive bool
	slowMotionTimer  *systems.Timer
	laserBeamActive  bool
	laserBeamTimer   *systems.Timer
	nukeActive       bool
	nukeTimer        *systems.Timer

	boss              *entities.Boss
	bossProjectiles   []*entities.BossProjectile
	bossBar           *ui.BossBar
	bossWarningShown  bool
	bossCooldownTimer *systems.Timer
	bossDefeated      bool
	bossCount         int
	bossNoDamage      bool

	meteorPool         *entities.MeteorPool
	laserPool          *entities.LaserPool
	powerUpPool        *entities.PowerUpPool
	bossProjectilePool *entities.BossProjectilePool

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
		slowMotionActive:   false,
		slowMotionTimer:    systems.NewTimer(config.SlowMotionTime),
		laserBeamActive:    false,
		laserBeamTimer:     systems.NewTimer(config.LaserBeamTime),
		nukeActive:         false,
		nukeTimer:          systems.NewTimer(config.NukeClearScreenTime),
		meteorPool:         entities.NewMeteorPool(),
		laserPool:          entities.NewLaserPool(),
		powerUpPool:        entities.NewPowerUpPool(),
		bossProjectilePool: entities.NewBossProjectilePool(),
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
		pauseIconX:         config.ScreenWidth - (config.PauseIconSize + config.PauseIconMargin),
		pauseIconY:         config.PauseIconMargin,
	}

	g.player = entities.NewPlayer(g)
	g.menu = ui.NewMenu()
	g.pauseMenu = ui.NewPauseMenu()
	g.initMobileControls()

	g.loadHighScore()
	g.loadLeaderboard()

	return g
}

func (g *Game) Update() error {
	g.updateStars()

	stateStart := time.Now()
	var err error
	switch g.state {
	case config.StateMenu:
		err = g.updateMenu()
	case config.StatePlaying:
		err = g.updatePlaying()
	case config.StateBossFight:
		err = g.updateBossFight()
	case config.StatePaused:
		err = g.updatePaused()
	case config.StateGameOver:
		err = g.updateGameOver()
	case config.StateWaitingNameInput:
		return nil
	}

	if time.Since(stateStart) > 10*time.Millisecond {
		fmt.Printf("⚠️ STATE UPDATE SLOW (%v): %v\n", g.state, time.Since(stateStart))
	}

	return err
}

func (g *Game) initMobileControls() {
	if g.joystick == nil {
		g.joystick = input.NewJoystick(100, float64(config.ScreenHeight-120), 60)
	}
	if g.shootButton == nil {
		g.shootButton = input.NewShootButton(float64(config.ScreenWidth-100), float64(config.ScreenHeight-120), 50)
	}
}

func (g *Game) updateMenu() error {
	g.menu.SetScores(g.highScore, g.lastScore)

	var touchIDs []ebiten.TouchID
	touchIDs = ebiten.AppendTouchIDs(touchIDs)

	if len(touchIDs) > 0 && !g.touchDetected {
		g.touchDetected = true
		g.isMobile = true
		g.initMobileControls()
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
		g.initMobileControls()
	}
	g.handleMobileControls(touchIDs)

	g.player.Update()
	g.notification.Update()

	speedMultiplier := 1.0 + float64(g.wave-1)*config.WaveDifficultyFactor
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

	g.updateGameTimers()

	for _, p := range g.powerUps {
		p.Update()
	}

	g.updateMeteors()

	for _, l := range g.lasers {
		l.Update()
	}

	for _, p := range g.particles {
		p.Update()
	}

	if g.boss != nil {
		g.boss.Update()
		for _, bp := range g.bossProjectiles {
			bp.Update()
		}
		for _, m := range g.boss.GetMinions() {
			if m != nil {
				m.Update()
			}
		}
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
		g.state = config.StatePlaying
	case ui.PauseActionRestart:
		g.Reset()
		g.state = config.StatePlaying
	case ui.PauseActionQuit:
		g.returnToMenu()
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyP) {
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
	g.updateAndSpawn(g.starSpawnTimer, func() {
		g.stars = append(g.stars, entities.NewStar())
	})

	for _, s := range g.stars {
		s.Update()
	}

	g.cleanStars()
}

func (g *Game) checkCollisions() bool {
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
				g.createExplosion(g.meteors[i].GetPosition(), config.ParticleCount)

				meteorsToRemove[i] = true
				if !g.laserBeamActive {
					lasersToRemove[j] = true
				}

				g.combo++
				g.comboTimer.Reset()
				g.addScore(1)

				assets.PlayExplosionSound()

				if !g.laserBeamActive {
					break
				}
			}
		}
	}

	g.filterMeteors(meteorsToRemove)
	g.filterLasers(lasersToRemove)

	for i := len(g.meteors) - 1; i >= 0; i-- {
		if g.meteors[i].Collider().Intersects(g.player.Collider()) {
			isDead := g.player.TakeDamage()

			g.meteorPool.Put(g.meteors[i])
			g.meteors = append(g.meteors[:i], g.meteors[i+1:]...)

			if isDead {
				return g.handleGameOver()
			}
			g.addScreenShake(config.ScreenShakeDuration)
			break
		}
	}

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

	return false
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

	g.drawParticlesBatch(screen)

	g.drawUI(screen)
	g.notification.Draw(screen)
}

func (g *Game) drawUI(screen *ebiten.Image) {
	lives := g.player.GetLives()

	for i := 0; i < lives && i < config.InitialLives; i++ {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(10+i*40), 10)
		screen.DrawImage(assets.HeartUISprite, op)
	}

	extraLives := max(0, lives-config.InitialLives)
	for i := 0; i < extraLives; i++ {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(10+(config.InitialLives+i)*40), 10)
		screen.DrawImage(assets.ExtraLifeUISprite, op)
	}

	waveText := fmt.Sprintf("Wave: %d", g.wave)
	drawText(screen, waveText, assets.FontSmall, 20, 65, color.White)

	if g.combo > 1 {
		comboText := fmt.Sprintf("%d COMBO", g.combo)
		comboColor := color.RGBA{255, 200, 0, 255}
		drawText(screen, comboText, assets.FontSmall, 20, 105, comboColor)
	}

	scoreText := fmt.Sprintf("Points: %d", g.score)
	drawText(screen, scoreText, assets.FontSmall, 20, 570, color.White)

	highScoreText := fmt.Sprintf("HIGH SCORE: %d", g.highScore)
	highScoreWidth := measureText(highScoreText, assets.FontSmall)
	highScoreX := config.ScreenWidth - highScoreWidth - 20
	drawText(screen, highScoreText, assets.FontSmall, highScoreX, 570, color.White)

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
		barY += 30
	}

	if g.slowMotionActive {
		progress := float32(g.slowMotionTimer.Progress())
		ui.DrawPowerUpBarAt(screen, progress, color.RGBA{100, 255, 255, 255}, barY)
		barY += 30
	}

	if g.laserBeamActive {
		progress := float32(g.laserBeamTimer.Progress())
		ui.DrawPowerUpBarAt(screen, progress, color.RGBA{150, 100, 255, 255}, barY)
		barY += 30
	}

	if g.nukeActive {
		progress := float32(g.nukeTimer.Progress())
		ui.DrawPowerUpBarAt(screen, progress, color.RGBA{255, 50, 50, 255}, barY)
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

	for _, pu := range g.powerUps {
		pu.Draw(screen)
	}

	g.drawParticlesBatch(screen)

	g.drawUI(screen)

	if g.boss != nil {
		g.bossBar.Draw(screen, g.boss.GetHealth(), g.boss.GetMaxHealth())
	}

	g.notification.Draw(screen)
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
	drawText(screen, scoreText, assets.FontSmall, 20, 570, color.White)

	highScoreText := fmt.Sprintf("HIGH SCORE: %d", g.highScore)
	highScoreWidth := measureText(highScoreText, assets.FontSmall)
	highScoreX := config.ScreenWidth - highScoreWidth - 20
	drawText(screen, highScoreText, assets.FontSmall, highScoreX, 570, color.White)
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
	g.screenShake = 0
	g.boss = nil
	g.bossDefeated = false
	g.bossWarningShown = false
	g.bossNoDamage = false
	g.bossCount = 0
	g.initNewGameSession()
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

// handlePowerUpCollected processa o efeito do power-up coletado
func (g *Game) handlePowerUpCollected(powerType entities.PowerUpType) {
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
	case entities.PowerUpSlowMotion:
		g.slowMotionActive = true
		g.slowMotionTimer.Reset()
		g.notification.Show("SLOW MOTION!", ui.NotificationSuperPower)
	case entities.PowerUpLaser:
		g.laserBeamActive = true
		g.laserBeamTimer.Reset()
		g.notification.Show("LASER BEAM!", ui.NotificationSuperPower)
	case entities.PowerUpNuke:
		g.activateNuke()
		g.notification.Show("NUKE ACTIVATED!", ui.NotificationSuperPower)
	case entities.PowerUpExtraLife:
		g.player.GainExtraLife()
		g.notification.Show("EXTRA LIFE!", ui.NotificationLife)
	}
}

func (g *Game) createExplosion(pos systems.Vector, count int) {
	for i := 0; i < count; i++ {
		g.particles = append(g.particles, effects.NewParticle(pos))
	}
}

func (g *Game) addScreenShake(intensity int) {
	g.screenShake = intensity
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

func (g *Game) handleGameOver() bool {
	g.saveHighScore()
	if g.leaderboard.IsTopScore(g.score) && g.hasNameInputModal() {
		g.state = config.StateWaitingNameInput
		g.showNameInputModal()
	} else {
		g.state = config.StateGameOver
	}
	return true
}

func (g *Game) filterMeteors(toRemove map[int]bool) {
	newMeteors := make([]*entities.Meteor, 0, len(g.meteors))
	for i, m := range g.meteors {
		if toRemove[i] {
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

func (g *Game) addScore(basePoints int) {
	points := basePoints + int(float64(g.combo)*config.ComboMultiplier)
	if points < 1 {
		points = 1
	}
	g.score += points

	if g.score >= g.wave*config.WaveScoreThreshold {
		g.wave++
	}
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
		(g.score >= config.BossScoreThreshold && g.score < config.BossScoreThreshold+config.BossScoreProximity && !g.bossWarningShown)
}

func (g *Game) spawnBoss() {
	g.bossWarningShown = true
	g.screenShake = config.BossWarningShakeTime
	g.notification.Show("BOSS APPROACHING!", ui.NotificationWarning)
	assets.PlayExplosionSound()

	bossType := config.BossType(g.bossCount % config.BossTypesCount)
	g.boss = entities.NewBoss(bossType)
	g.bossNoDamage = true
	g.bossBar.Show()
	g.state = config.StateBossFight
	g.powerUpSpawnTimer = systems.NewTimer(config.PowerUpSpawnTimeBoss)
}

func (g *Game) updateBossFight() error {
	if g.shouldPause() {
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

		if g.boss.CanShoot() && g.boss.GetPosition().Y >= 100 {
			g.boss.Shoot()
			pos := g.boss.GetPosition()
			bp := g.bossProjectilePool.Get()
			bp.Reset(pos.X, pos.Y+40)
			g.bossProjectiles = append(g.bossProjectiles, bp)
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

		if g.boss.GetBossType() == config.BossSwarm {
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
						g.boss.RemoveMinion(mIdx)
					}
					break
				}
			}
		}
	}

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

func (g *Game) drawParticlesBatch(screen *ebiten.Image) {
	if len(g.particles) == 0 {
		return
	}

	vertices := make([]ebiten.Vertex, 0, len(g.particles)*4)
	indices := make([]uint16, 0, len(g.particles)*6)

	for _, p := range g.particles {
		x := float32(p.GetPosition().X)
		y := float32(p.GetPosition().Y)
		size := float32(2.0)
		col := p.GetColor()

		r := float32(col.R) / 255.0
		gVal := float32(col.G) / 255.0
		b := float32(col.B) / 255.0
		a := float32(col.A) / 255.0

		baseIdx := uint16(len(vertices))

		vertices = append(vertices,
			ebiten.Vertex{DstX: x - size, DstY: y - size, SrcX: 0, SrcY: 0, ColorR: r, ColorG: gVal, ColorB: b, ColorA: a},
			ebiten.Vertex{DstX: x + size, DstY: y - size, SrcX: 1, SrcY: 0, ColorR: r, ColorG: gVal, ColorB: b, ColorA: a},
			ebiten.Vertex{DstX: x + size, DstY: y + size, SrcX: 1, SrcY: 1, ColorR: r, ColorG: gVal, ColorB: b, ColorA: a},
			ebiten.Vertex{DstX: x - size, DstY: y + size, SrcX: 0, SrcY: 1, ColorR: r, ColorG: gVal, ColorB: b, ColorA: a},
		)

		indices = append(indices,
			baseIdx, baseIdx+1, baseIdx+2,
			baseIdx, baseIdx+2, baseIdx+3,
		)
	}

	screen.DrawTriangles(vertices, indices, emptyImage, nil)
}
