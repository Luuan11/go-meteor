package core

import (
	"fmt"
	"image/color"

	"go-meteor/internal/config"
	"go-meteor/internal/ui"
	assets "go-meteor/src/pkg"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
)

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
	case config.StateBossAnnouncement:
		g.drawPlaying(screen)
		g.drawBossAnnouncement(screen)
	case config.StateBossFight:
		g.drawBossFight(screen)
	case config.StatePaused:
		g.drawPlaying(screen)
		g.pauseMenu.Draw(screen)
	case config.StateGameOver:
		g.drawGameOver(screen)
	case config.StateShop:
		g.drawGameOver(screen)
		g.shop.Draw(screen)
	case config.StateSettings:
		g.drawMenu(screen)
		g.settingsMenu.Draw(screen)
	case config.StatePlayerDeath:
		g.drawPlayerDeath(screen)
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

	for _, c := range g.coins {
		c.Draw(screen)
	}

	g.drawParticlesBatch(screen)

	g.drawUI(screen)
	g.notification.Draw(screen)
}

func (g *Game) drawUI(screen *ebiten.Image) {
	lives := g.player.GetLives()

	for i := 0; i < lives && i < config.InitialLives; i++ {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(config.HeartOffsetX+i*config.HeartSpacing), config.HeartOffsetY)
		screen.DrawImage(assets.HeartUISprite, op)
	}

	extraLives := max(0, lives-config.InitialLives)
	for i := 0; i < extraLives; i++ {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(config.HeartOffsetX+(config.InitialLives+i)*config.HeartSpacing), config.HeartOffsetY)
		screen.DrawImage(assets.ExtraLifeUISprite, op)
	}

	waveText := fmt.Sprintf("Wave: %d", g.wave)
	drawText(screen, waveText, assets.FontSmall, 20, 65, color.White)

	if g.progress != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(1, 1)
		op.GeoM.Translate(20, 75)
		screen.DrawImage(assets.CoinSprite, op)

		coinsText := fmt.Sprintf("%d", g.progress.Coins)
		drawText(screen, coinsText, assets.FontSmall, 60, 100, color.RGBA{255, 215, 0, 255})
	}

	if g.combo > 1 {
		comboText := fmt.Sprintf("%d COMBO", g.combo)
		comboColor := color.RGBA{255, 200, 0, 255}
		drawText(screen, comboText, assets.FontSmall, 20, 130, comboColor)
	}

	scoreText := fmt.Sprintf("Points: %d", g.score)
	drawText(screen, scoreText, assets.FontSmall, 20, 570, color.White)

	highScoreText := fmt.Sprintf("HIGH SCORE: %d", g.highScore)
	highScoreWidth := measureText(highScoreText, assets.FontSmall)
	highScoreX := config.ScreenWidth - highScoreWidth - 20
	drawText(screen, highScoreText, assets.FontSmall, highScoreX, 570, color.White)

	ui.DrawPauseIcon(screen, g.pauseIconX, g.pauseIconY)

	barY := float32(config.PowerUpBarStartY)

	if g.superPowerActive {
		progress := float32(g.superPowerTimer.Progress())
		ui.DrawPowerUpBar(screen, progress, color.RGBA{255, 100, 255, 255})
		barY += config.PowerUpBarSpacing
	}

	if g.player.HasShield() {
		progress := float32(g.player.ShieldProgress())
		ui.DrawPowerUpBarAt(screen, progress, color.RGBA{100, 200, 255, 255}, barY)
		barY += config.PowerUpBarSpacing
	}

	if g.slowMotionActive {
		progress := float32(g.slowMotionTimer.Progress())
		ui.DrawPowerUpBarAt(screen, progress, color.RGBA{100, 255, 255, 255}, barY)
		barY += config.PowerUpBarSpacing
	}

	if g.laserBeamActive {
		progress := float32(g.laserBeamTimer.Progress())
		ui.DrawPowerUpBarAt(screen, progress, color.RGBA{150, 100, 255, 255}, barY)
		barY += config.PowerUpBarSpacing
	}

	if g.nukeActive {
		progress := float32(g.nukeTimer.Progress())
		ui.DrawPowerUpBarAt(screen, progress, color.RGBA{255, 50, 50, 255}, barY)
		barY += 30
	}

	if g.multiplierActive {
		progress := float32(g.multiplierTimer.Progress())
		ui.DrawPowerUpBarAt(screen, progress, color.RGBA{255, 215, 0, 255}, barY)
	}

	if g.isMobile && g.joystick != nil && g.shootButton != nil {
		g.joystick.Draw(screen)
		g.shootButton.Draw(screen)
	}
}

func (g *Game) drawBossAnnouncement(screen *ebiten.Image) {
	g.player.Draw(screen)

	for _, m := range g.meteors {
		m.Draw(screen)
	}

	for _, b := range g.lasers {
		b.Draw(screen)
	}

	for _, p := range g.particles {
		p.Draw(screen)
	}

	g.drawUI(screen)

	alpha := uint8(255)
	if g.bossAnnouncementTimer < config.BossAnnouncementFade {
		alpha = uint8(float64(g.bossAnnouncementTimer) / float64(config.BossAnnouncementFade) * 255)
	}

	txt := "BOSS INCOMING!"
	face := assets.FontUi
	txtWidth := measureText(txt, face)
	x := (config.ScreenWidth - txtWidth) / 2
	y := config.ScreenHeight / 2

	shadowOffset := 4
	shadowColor := color.RGBA{0, 0, 0, alpha}
	drawText(screen, txt, face, x+shadowOffset, y+shadowOffset, shadowColor)

	textColor := color.RGBA{255, 50, 50, alpha}
	drawText(screen, txt, face, x, y, textColor)

	if g.isMobile && g.joystick != nil && g.shootButton != nil {
		g.joystick.Draw(screen)
		g.shootButton.Draw(screen)
	}
}

func (g *Game) drawBossFight(screen *ebiten.Image) {
	g.player.Draw(screen)

	if g.boss != nil {
		g.boss.Draw(screen)

		// Draw minions
		minions := g.boss.GetMinions()
		for i, minion := range minions {
			if minion != nil {
				minion.Draw(screen)
			}
			_ = i // debug: usar i para ver quantos minions existem
		}
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
	youDiedText := "GAME OVER"
	youDiedWidth := measureText(youDiedText, assets.FontUi)
	youDiedX := (config.ScreenWidth - youDiedWidth) / 2
	drawText(screen, youDiedText, assets.FontUi, youDiedX, 150, color.White)

	if g.statistics != nil {
		g.statistics.Draw(screen, 220)
	}

	tryAgainText := "Press ENTER to try again"
	tryAgainWidth := measureText(tryAgainText, assets.FontUi)
	tryAgainX := (config.ScreenWidth - tryAgainWidth) / 2
	drawText(screen, tryAgainText, assets.FontUi, tryAgainX, 480, color.White)

	if g.isMobile {
		shopButtonWidth := 180
		shopButtonHeight := 40
		shopButtonX := (config.ScreenWidth - shopButtonWidth) / 2
		shopButtonY := 535

		shopButton := ebiten.NewImage(shopButtonWidth, shopButtonHeight)
		shopButton.Fill(color.RGBA{255, 215, 0, 255})
		shopButtonOp := &ebiten.DrawImageOptions{}
		shopButtonOp.GeoM.Translate(float64(shopButtonX), float64(shopButtonY))
		screen.DrawImage(shopButton, shopButtonOp)

		shopButtonBorder := ebiten.NewImage(shopButtonWidth, shopButtonHeight)
		for y := 0; y < shopButtonHeight; y++ {
			for x := 0; x < shopButtonWidth; x++ {
				if x < 3 || x >= shopButtonWidth-3 || y < 3 || y >= shopButtonHeight-3 {
					shopButtonBorder.Set(x, y, color.RGBA{200, 170, 0, 255})
				}
			}
		}
		screen.DrawImage(shopButtonBorder, shopButtonOp)

		shopText := "SHOP"
		shopTextWidth := measureText(shopText, assets.FontSmall)
		shopTextX := shopButtonX + (shopButtonWidth-shopTextWidth)/2
		drawText(screen, shopText, assets.FontSmall, shopTextX, shopButtonY+12, color.RGBA{0, 0, 0, 255})
	} else {
		shopText := "Press S to open SHOP"
		shopWidth := measureText(shopText, assets.FontSmall)
		shopX := (config.ScreenWidth - shopWidth) / 2
		drawText(screen, shopText, assets.FontSmall, shopX, 535, color.RGBA{255, 215, 0, 255})
	}

	highScoreText := fmt.Sprintf("HIGH SCORE: %d", g.highScore)
	highScoreWidth := measureText(highScoreText, assets.FontSmall)
	highScoreX := config.ScreenWidth - highScoreWidth - 20
	drawText(screen, highScoreText, assets.FontSmall, highScoreX, 570, color.RGBA{255, 215, 0, 255})
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

func (g *Game) drawPlayerDeath(screen *ebiten.Image) {
	// Draw game background (meteors, stars, particles)
	for _, s := range g.stars {
		s.Draw(screen)
	}

	for _, m := range g.meteors {
		m.Draw(screen)
	}

	for _, p := range g.particles {
		p.Draw(screen)
	}

	// Don't draw player (they exploded)
	// Draw coins and power-ups
	for _, c := range g.coins {
		c.Draw(screen)
	}

	for _, pu := range g.powerUps {
		pu.Draw(screen)
	}

	// Draw UI
	scoreText := fmt.Sprintf("Score: %d", g.score)
	drawText(screen, scoreText, assets.FontSmall, 20, 30, color.White)

	waveText := fmt.Sprintf("Wave: %d", g.wave)
	drawText(screen, waveText, assets.FontSmall, 20, 65, color.White)
}

func drawText(screen *ebiten.Image, txt string, face font.Face, x, y int, clr color.Color) {
	text.Draw(screen, txt, face, x, y, clr)
}

func measureText(txt string, face font.Face) int {
	bounds := text.BoundString(face, txt)
	return bounds.Dx()
}
