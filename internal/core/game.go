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

	meteorSpawnTimer  *systems.Timer
	planetSpawnTimer  *systems.Timer
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
	planets          []*entities.Planet
	lasers           []*entities.Laser
	powerUps         []*entities.PowerUp
	particles        []*effects.Particle
	superPowerActive bool

	meteorPool  *entities.MeteorPool
	laserPool   *entities.LaserPool
	powerUpPool *entities.PowerUpPool

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
}

func NewGame() *Game {
	g := &Game{
		state:             config.StateMenu,
		meteorSpawnTimer:  systems.NewTimer(config.MeteorSpawnTime),
		starSpawnTimer:    systems.NewTimer(config.StarSpawnTime),
		planetSpawnTimer:  systems.NewTimer(config.PlanetSpawnTime),
		powerUpSpawnTimer: systems.NewTimer(config.PowerUpSpawnTime),
		superPowerTimer:   systems.NewTimer(config.SuperPowerTime),
		comboTimer:        systems.NewTimer(config.ComboTimeout),
		superPowerActive:  false,
		meteorPool:        entities.NewMeteorPool(),
		laserPool:         entities.NewLaserPool(),
		powerUpPool:       entities.NewPowerUpPool(),
		notification:      ui.NewNotification(),
		wave:              1,
		isMobile:          false,
		touchDetected:     false,
		leaderboard:       systems.NewLeaderboard(),
		storage:           systems.NewStorage(),
	}

	g.joystick = input.NewJoystick(100, float64(config.ScreenHeight-120), 60)
	g.shootButton = input.NewShootButton(float64(config.ScreenWidth-100), float64(config.ScreenHeight-120), 50)
	g.pauseIconX = config.ScreenWidth - 45
	g.pauseIconY = 15

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
	var touchIDs []ebiten.TouchID
	touchIDs = ebiten.AppendTouchIDs(touchIDs)

	if len(touchIDs) > 0 && !g.touchDetected {
		g.touchDetected = true
		g.isMobile = true
	}

	g.menu.Update()

	if g.menu.IsReady() {
		g.planets = nil
		g.initNewGameSession()
		g.state = config.StatePlaying
	}

	g.updateAndSpawn(g.planetSpawnTimer, func() {
		g.planets = append(g.planets, entities.NewPlanet())
	})

	for _, m := range g.planets {
		m.Update()
	}

	g.cleanPlanets()
	return nil
}

func (g *Game) updatePlaying() error {
	if g.shouldPause() {
		assets.PauseMusic()
		g.state = config.StatePaused
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

	g.updateAndSpawn(g.meteorSpawnTimer, func() {
		for i := 0; i < meteorsPerWave; i++ {
			m := g.meteorPool.Get()
			m.Reset(speedMultiplier)
			g.meteors = append(g.meteors, m)
		}
	})

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
	g.updateAndSpawn(g.starSpawnTimer, func() {
		g.stars = append(g.stars, entities.NewStar())
	})

	for _, m := range g.stars {
		m.Update()
	}

	g.cleanStars()
}

func (g *Game) checkCollisions() bool {
	for i := len(g.meteors) - 1; i >= 0; i-- {
		for j := len(g.lasers) - 1; j >= 0; j-- {
			if g.meteors[i].Collider().Intersects(g.lasers[j].Collider()) {
				pos := g.meteors[i].GetPosition()

				for k := 0; k < config.ParticleCount; k++ {
					g.particles = append(g.particles, effects.NewParticle(pos))
				}

				g.meteorPool.Put(g.meteors[i])
				g.meteors = append(g.meteors[:i], g.meteors[i+1:]...)

				g.laserPool.Put(g.lasers[j])
				g.lasers = append(g.lasers[:j], g.lasers[j+1:]...)

				g.combo++
				g.comboTimer.Reset()
				points := 1 + int(float64(g.combo)*config.ComboMultiplier)
				g.score += points

				assets.PlayExplosionSound()

				if g.score >= g.wave*config.WaveScoreThreshold {
					g.wave++
				}

				break
			}
		}
	}

	for i := len(g.meteors) - 1; i >= 0; i-- {
		if g.meteors[i].Collider().Intersects(g.player.Collider()) {
			isDead := g.player.TakeDamage()

			g.meteorPool.Put(g.meteors[i])
			g.meteors = append(g.meteors[:i], g.meteors[i+1:]...)

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

	for i, p := range g.powerUps {
		if p.Collider().Intersects(g.player.Collider()) {
			powerType := p.GetType()
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
	for i := len(g.meteors) - 1; i >= 0; i-- {
		if g.meteors[i].IsOutOfScreen() {
			g.meteorPool.Put(g.meteors[i])
			g.meteors = append(g.meteors[:i], g.meteors[i+1:]...)
		}
	}

	for i := len(g.lasers) - 1; i >= 0; i-- {
		if g.lasers[i].IsOutOfScreen() {
			g.laserPool.Put(g.lasers[i])
			g.lasers = append(g.lasers[:i], g.lasers[i+1:]...)
		}
	}

	for i := len(g.powerUps) - 1; i >= 0; i-- {
		if g.powerUps[i].IsOutOfScreen() {
			g.powerUpPool.Put(g.powerUps[i])
			g.powerUps = append(g.powerUps[:i], g.powerUps[i+1:]...)
		}
	}

	for i := len(g.particles) - 1; i >= 0; i-- {
		if g.particles[i].IsDead() {
			g.particles = append(g.particles[:i], g.particles[i+1:]...)
		}
	}
}

func (g *Game) cleanStars() {
	for i := len(g.stars) - 1; i >= 0; i-- {
		if g.stars[i].IsOutOfScreen() {
			g.stars = append(g.stars[:i], g.stars[i+1:]...)
		}
	}
}

func (g *Game) cleanPlanets() {
	for i := len(g.planets) - 1; i >= 0; i-- {
		if g.planets[i].IsOutOfScreen() {
			g.planets = append(g.planets[:i], g.planets[i+1:]...)
		}
	}
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

	for _, b := range g.stars {
		b.Draw(screen)
	}

	switch g.state {
	case config.StateMenu:
		g.drawMenu(screen)
	case config.StatePlaying:
		g.drawPlaying(screen)
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
	for _, b := range g.planets {
		b.Draw(screen)
	}

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

	if g.isMobile {
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
	if !g.isMobile {
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
}

func (g *Game) returnToMenu() {
	g.clearPools()
	g.meteors = nil
	g.lasers = nil
	g.powerUps = nil
	g.particles = nil
	g.player = entities.NewPlayer(g)
	g.Reset()
	g.menu.Reset()
	g.state = config.StateMenu
}

func (g *Game) Reset() {
	g.clearPools()

	g.player = entities.NewPlayer(g)
	g.meteors = nil
	g.lasers = nil
	g.powerUps = nil
	g.particles = nil
	g.meteorSpawnTimer.Reset()
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
