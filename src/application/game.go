package game

import (
	"fmt"
	assets "go-meteor/src/pkg"
	menu "go-meteor/src/resources"
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
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

	for i, m := range g.meteors {
		for j, b := range g.lasers {
			if m.Collider().Intersects(b.Collider()) && (i >= 0 && i < len(g.meteors)) {
				g.meteors = append(g.meteors[:i], g.meteors[i+1:]...)
				g.lasers = append(g.lasers[:j], g.lasers[j+1:]...)
				g.score++
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
