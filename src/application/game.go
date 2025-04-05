package game

import (
	"fmt"
	assets "go-meteor/src/pkg"
	menu "go-meteor/src/resources"
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text"
)

var bestScore = 0

const (
	screenWidth      = 800
	screenHeight     = 600
	meteorSpawnTime  = 1 * time.Second
	starSpawnTime    = (1 * time.Second) / 2
	planetSpawnTime  = 5 * time.Second
	powerUpSpawnTime = 20 * time.Second
	superPowerTime   = 10 * time.Second
)

type Game struct {
	meteorSpawnTimer  *Timer
	planetSpawnTimer  *Timer
	starSpawnTimer    *Timer
	powerUpSpawnTimer *Timer
	superPowerTimer   *Timer
	menu              *menu.Menu

	player           *Player
	meteors          []*Meteor
	stars            []*Star
	planets          []*Planet
	lasers           []*Laser
	powerUps         []*PowerUp
	superPowerActive bool

	isStarted  bool
	score      int
	isGameOver bool
}

func NewGame() *Game {
	g := &Game{
		meteorSpawnTimer:  NewTimer(meteorSpawnTime),
		starSpawnTimer:    NewTimer(starSpawnTime),
		planetSpawnTimer:  NewTimer(planetSpawnTime),
		powerUpSpawnTimer: NewTimer(powerUpSpawnTime),
		superPowerTimer:   NewTimer(superPowerTime),
		superPowerActive:  false,
	}

	g.player = NewPlayer(g)
	g.menu = menu.NewMenu()

	return g
}

func (g *Game) Update() error {
	if g.isGameOver {
		if ebiten.IsKeyPressed(ebiten.KeyEnter) {
			g.isGameOver = false
			g.isStarted = false
		}
		return nil
	}

	if g.isGameOver {
		var touchIDs []ebiten.TouchID
		touchIDs = ebiten.AppendTouchIDs(touchIDs)
		if len(touchIDs) > 0 {
			g.Reset()
			g.isStarted = true
			return nil
		}
	}

	g.starSpawnTimer.Update()
	if g.starSpawnTimer.IsReady() {
		g.starSpawnTimer.Reset()

		s := NewStar()
		g.stars = append(g.stars, s)
	}

	for _, m := range g.stars {
		m.Update()
	}

	if !g.isStarted {

		g.menu.Update()

		if g.menu.IsReady() {
			g.planets = nil
			g.isStarted = true
		}

		g.planetSpawnTimer.Update()
		if g.planetSpawnTimer.IsReady() {
			g.planetSpawnTimer.Reset()

			s := NewPlanet()
			g.planets = append(g.planets, s)
		}

		for _, m := range g.planets {
			m.Update()
		}

		return nil
	}

	g.player.Update()

	g.meteorSpawnTimer.Update()
	if g.meteorSpawnTimer.IsReady() {
		g.meteorSpawnTimer.Reset()

		m := NewMeteor()
		g.meteors = append(g.meteors, m)
	}

	g.powerUpSpawnTimer.Update()
	if g.powerUpSpawnTimer.IsReady() {
		g.powerUpSpawnTimer.Reset()

		p := NewPowerUp()
		g.powerUps = append(g.powerUps, p)
	}

	if g.superPowerActive {
		g.superPowerTimer.Update()
		if g.superPowerTimer.IsReady() {
			g.superPowerTimer.Reset()
			g.superPowerActive = false
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

	for i := len(g.meteors) - 1; i >= 0; i-- {
		for j := len(g.lasers) - 1; j >= 0; j-- {
			if g.meteors[i].Collider().Intersects(g.lasers[j].Collider()) {
				g.meteors = append(g.meteors[:i], g.meteors[i+1:]...)
				g.lasers = append(g.lasers[:j], g.lasers[j+1:]...)
				g.score++
				break
			}
		}
	}

	for _, m := range g.meteors {
		if m.Collider().Intersects(g.player.Collider()) {
			g.Reset()
			break
		}
	}

	for i, p := range g.powerUps {
		if p.Collider().Intersects(g.player.Collider()) {
			g.powerUps = append(g.powerUps[:i], g.powerUps[i+1:]...)
			g.superPowerActive = true
			g.superPowerTimer.Reset()
			break
		}
	}

	var touchIDs []ebiten.TouchID
	touchIDs = ebiten.AppendTouchIDs(touchIDs)

	for _, id := range touchIDs {
		x, y := ebiten.TouchPosition(id)

		if x >= 50 && x <= 50+80 && y >= screenHeight-250 && y <= screenHeight-250+80 {
			g.player.MoveLeft()
		}

		if x >= 190 && x <= 190+80 && y >= screenHeight-250 && y <= screenHeight-250+80 {
			g.player.MoveRight()
		}

		if x >= 120 && x <= 120+80 && y >= screenHeight-330 && y <= screenHeight-330+80 {
			g.player.MoveUp()
		}

		if x >= 120 && x <= 120+80 && y >= screenHeight-170 && y <= screenHeight-170+80 {
			g.player.MoveDown()
		}

		if x >= screenWidth-150 && x <= screenWidth-150+80 && y >= screenHeight-250 && y <= screenHeight-250+80 {
			g.player.Shoot()
		}
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	if g.isGameOver {
		youDiedText := "YOU DIED"
		youDiedBounds := text.BoundString(assets.FontUi, youDiedText)
		youDiedX := (screenWidth - youDiedBounds.Dx()) / 2
		text.Draw(screen, youDiedText, assets.FontUi, youDiedX, 300, color.White)

		tryAgainText := "Press Enter to try again"
		tryAgainBounds := text.BoundString(assets.FontUi, tryAgainText)
		tryAgainX := (screenWidth - tryAgainBounds.Dx()) / 2
		text.Draw(screen, tryAgainText, assets.FontUi, tryAgainX, 400, color.White)

		text.Draw(screen, fmt.Sprintf("Points: %d         High Score: %d", g.score, bestScore), assets.FontUi, 20, 570, color.White)
		return
	}

	for _, b := range g.stars {
		b.Draw(screen)
	}

	if !g.isStarted {
		for _, b := range g.planets {
			b.Draw(screen)
		}

		g.menu.Draw(screen)
		return
	}

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

	scoreSprite := &ebiten.DrawImageOptions{}

	scoreSprite.GeoM.Translate(60, 450)

	text.Draw(screen, fmt.Sprintf("Points: %d            High Score: %d", g.score, bestScore), assets.FontUi, 20, 570, color.White)

	drawButton(screen, "◀", 50, screenHeight-250)              // Esquerda
	drawButton(screen, "▶", 190, screenHeight-250)             // Direita
	drawButton(screen, "▲", 120, screenHeight-330)             // Cima
	drawButton(screen, "▼", 120, screenHeight-170)             // Baixo
	drawButton(screen, "●", screenWidth-150, screenHeight-250) // Outro botão (OK/ação/etc.)
}

func drawButton(screen *ebiten.Image, label string, x, y int) {
	btnWidth, btnHeight := 80, 80
	
	ebitenutil.DrawRect(screen, float64(x), float64(y), float64(btnWidth), float64(btnHeight), color.RGBA{0, 0, 0, 100})

	textBounds := text.BoundString(assets.ScoreFont, label)
	textWidth := textBounds.Dx()
	textHeight := textBounds.Dy()
	textX := x + (btnWidth-textWidth)/2
	textY := y + (btnHeight-textHeight)/2 + textHeight

	text.Draw(screen, label, assets.FontBtn, textX, textY, color.White)
}

func (g *Game) AddLaser(l *Laser) {
	g.lasers = append(g.lasers, l)
}

func (g *Game) Reset() {
	g.player = NewPlayer(g)
	g.meteors = nil
	g.lasers = nil
	g.powerUps = nil
	g.meteorSpawnTimer.Reset()
	g.starSpawnTimer.Reset()
	g.powerUpSpawnTimer.Reset()

	if g.score >= bestScore {
		bestScore = g.score
	}

	g.score = 0
	g.superPowerActive = false
	g.isGameOver = true
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}
