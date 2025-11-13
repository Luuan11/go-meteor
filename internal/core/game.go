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

	lifeNotificationTimer int
	showLifeNotification  bool

	leaderboard *systems.Leaderboard
	storage     systems.Storage
	highScore   int

	isTopScore bool
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
	assets.UpdateAudio()

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
		g.state = config.StatePlaying
	}

	g.planetSpawnTimer.Update()
	if g.planetSpawnTimer.IsReady() {
		g.planetSpawnTimer.Reset()
		s := entities.NewPlanet()
		g.planets = append(g.planets, s)
	}

	for _, m := range g.planets {
		m.Update()
	}

	g.cleanPlanets()
	return nil
}

func (g *Game) updatePlaying() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyP) {
		g.state = config.StatePaused
		return nil
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		if x >= g.pauseIconX && x <= g.pauseIconX+30 && y >= g.pauseIconY && y <= g.pauseIconY+30 {
			g.state = config.StatePaused
			return nil
		}
	}

	var touchIDs []ebiten.TouchID
	touchIDs = ebiten.AppendTouchIDs(touchIDs)

	for _, id := range touchIDs {
		x, y := ebiten.TouchPosition(id)
		if x >= g.pauseIconX && x <= g.pauseIconX+30 && y >= g.pauseIconY && y <= g.pauseIconY+30 {
			g.state = config.StatePaused
			return nil
		}
	}

	if len(touchIDs) > 0 && !g.touchDetected {
		g.touchDetected = true
		g.isMobile = true
	}

	if g.isMobile {
		g.joystick.Update(touchIDs)
		if g.shootButton.Update(touchIDs) {
			g.player.Shoot()
		}

		dx, dy := g.joystick.GetDirection()
		if g.joystick.IsPressed() {
			if dx < -0.3 {
				g.player.MoveLeft()
			}
			if dx > 0.3 {
				g.player.MoveRight()
			}
			if dy < -0.3 {
				g.player.MoveUp()
			}
			if dy > 0.3 {
				g.player.MoveDown()
			}
		}
	}

	g.player.Update()

	speedMultiplier := 1.0 + float64(g.wave-1)*0.15
	meteorsPerWave := 1
	if g.wave > 1 {
		meteorsPerWave = 1 + (g.wave-1)/5
	}

	g.meteorSpawnTimer.Update()
	if g.meteorSpawnTimer.IsReady() {
		g.meteorSpawnTimer.Reset()

		for i := 0; i < meteorsPerWave; i++ {
			m := g.meteorPool.Get()
			m.Reset(speedMultiplier)
			g.meteors = append(g.meteors, m)
		}
	}

	g.powerUpSpawnTimer.Update()
	if g.powerUpSpawnTimer.IsReady() {
		g.powerUpSpawnTimer.Reset()

		p := g.powerUpPool.Get()
		p.Reset()
		g.powerUps = append(g.powerUps, p)
	}

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

	g.cleanObjects()

	if g.screenShake > 0 {
		g.screenShake--
	}

	if g.showLifeNotification {
		g.lifeNotificationTimer--
		if g.lifeNotificationTimer <= 0 {
			g.showLifeNotification = false
		}
	}

	return nil
}

func (g *Game) updatePaused() error {
	action := g.pauseMenu.Update()

	if action == 0 {
		g.state = config.StatePlaying
	} else if action == 1 {
		g.Reset()
		g.state = config.StatePlaying
	} else if action == 2 {
		for _, m := range g.meteors {
			g.meteorPool.Put(m)
		}
		for _, l := range g.lasers {
			g.laserPool.Put(l)
		}
		for _, p := range g.powerUps {
			g.powerUpPool.Put(p)
		}
		g.meteors = nil
		g.lasers = nil
		g.powerUps = nil
		g.particles = nil
		g.player = entities.NewPlayer(g)
		g.menu.Reset()
		g.state = config.StateMenu
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyP) {
		g.state = config.StatePlaying
	}

	return nil
}

func (g *Game) updateGameOver() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.Reset()
		g.menu.Reset()
		g.state = config.StateMenu
		return nil
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		g.Reset()
		g.menu.Reset()
		g.state = config.StateMenu
		return nil
	}

	var touchIDs []ebiten.TouchID
	touchIDs = inpututil.AppendJustPressedTouchIDs(touchIDs)
	if len(touchIDs) > 0 {
		g.Reset()
		g.menu.Reset()
		g.state = config.StateMenu
		return nil
	}

	return nil
}

func (g *Game) updateStars() {
	g.starSpawnTimer.Update()
	if g.starSpawnTimer.IsReady() {
		g.starSpawnTimer.Reset()
		s := entities.NewStar()
		g.stars = append(g.stars, s)
	}

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

	for _, m := range g.meteors {
		if m.Collider().Intersects(g.player.Collider()) {
			isDead := g.player.TakeDamage()
			if isDead {
				g.saveHighScore()
				if g.leaderboard.IsTopScore(g.score) && g.hasNameInputModal() {
					g.isTopScore = true
					g.state = config.StateWaitingNameInput
					g.showNameInputModal()
				} else {
					g.isTopScore = false
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
			case entities.PowerUpHeart:
				g.player.Heal()
				g.showLifeNotification = true
				g.lifeNotificationTimer = 120
			case entities.PowerUpShield:
				g.player.ActivateShield()
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
}

func (g *Game) drawUI(screen *ebiten.Image) {
	for i := 0; i < config.InitialLives; i++ {
		filled := i < g.player.GetLives()
		ui.DrawHeart(screen, 20+i*25, 20, filled)
	}

	waveText := fmt.Sprintf("Wave: %d", g.wave)
	text.Draw(screen, waveText, assets.FontSmall, 20, 65, color.White)

	if g.combo > 1 {
		comboText := fmt.Sprintf("%d COMBO", g.combo)
		comboColor := color.RGBA{255, 200, 0, 255}
		text.Draw(screen, comboText, assets.FontSmall, 20, 105, comboColor)
	}

	scoreText := fmt.Sprintf("Points: %d", g.score)
	text.Draw(screen, scoreText, assets.FontUi, 20, 570, color.White)

	highScoreText := fmt.Sprintf("HIGH SCORE: %d", g.highScore)
	highScoreWidth := font.MeasureString(assets.FontUi, highScoreText)
	highScoreX := config.ScreenWidth - highScoreWidth.Ceil() - 20
	text.Draw(screen, highScoreText, assets.FontUi, highScoreX, 570, color.White)

	ui.DrawPauseIcon(screen, g.pauseIconX, g.pauseIconY)

	if g.isMobile {
		g.joystick.Draw(screen)
		g.shootButton.Draw(screen)
	}
}

func (g *Game) drawGameOver(screen *ebiten.Image) {
	youDiedText := "YOU DIED"
	youDiedWidth := font.MeasureString(assets.FontUi, youDiedText)
	youDiedX := (config.ScreenWidth - youDiedWidth.Ceil()) / 2
	text.Draw(screen, youDiedText, assets.FontUi, youDiedX, 300, color.White)

	tryAgainText := "Press ENTER to try again"
	tryAgainWidth := font.MeasureString(assets.FontUi, tryAgainText)
	tryAgainX := (config.ScreenWidth - tryAgainWidth.Ceil()) / 2
	text.Draw(screen, tryAgainText, assets.FontUi, tryAgainX, 400, color.White)

	scoreText := fmt.Sprintf("Points: %d", g.score)
	text.Draw(screen, scoreText, assets.FontUi, 20, 570, color.White)

	highScoreText := fmt.Sprintf("High Score: %d", g.highScore)
	highScoreWidth := font.MeasureString(assets.FontUi, highScoreText)
	highScoreX := config.ScreenWidth - highScoreWidth.Ceil() - 20
	text.Draw(screen, highScoreText, assets.FontUi, highScoreX, 570, color.White)
}

func (g *Game) AddLaser(l *entities.Laser) {
	g.lasers = append(g.lasers, l)
}

func (g *Game) Reset() {
	for _, m := range g.meteors {
		g.meteorPool.Put(m)
	}
	for _, l := range g.lasers {
		g.laserPool.Put(l)
	}
	for _, p := range g.powerUps {
		g.powerUpPool.Put(p)
	}

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

func (g *Game) saveLeaderboard() {
	data, err := g.leaderboard.ToJSON()
	if err == nil {
		g.storage.SaveLeaderboard(data)
	}
}

func (g *Game) addScoreToLeaderboard(name string, score int) {
	// Add to local leaderboard
	g.leaderboard.AddScore(name, score)
	g.saveLeaderboard()

	// Also notify web leaderboard if running in browser
	g.notifyWebLeaderboard(name, score)
}
