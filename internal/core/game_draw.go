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
	g.drawLives(screen)
	g.drawWaveAndCoins(screen)
	g.drawScores(screen)
	ui.DrawPauseIcon(screen, g.pauseIconX, g.pauseIconY)
	g.drawPowerUpBars(screen)
	g.drawMobileControls(screen)
}

func (g *Game) drawLives(screen *ebiten.Image) {
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
}

func (g *Game) drawWaveAndCoins(screen *ebiten.Image) {
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
		drawText(screen, comboText, assets.FontSmall, 20, 130, color.RGBA{255, 200, 0, 255})
	}
}

func (g *Game) drawScores(screen *ebiten.Image) {
	scoreText := fmt.Sprintf("Points: %d", g.score)
	drawText(screen, scoreText, assets.FontSmall, 20, 570, color.White)

	highScoreText := fmt.Sprintf("HIGH SCORE: %d", g.highScore)
	highScoreX := config.ScreenWidth - measureText(highScoreText, assets.FontSmall) - 20
	drawText(screen, highScoreText, assets.FontSmall, highScoreX, 570, color.White)
}

func (g *Game) drawPowerUpBars(screen *ebiten.Image) {
	barY := float32(config.PowerUpBarStartY)

	if g.superPowerActive {
		ui.DrawPowerUpBarAt(screen, float32(g.superPowerTimer.Progress()), color.RGBA{255, 100, 255, 255}, barY)
		barY += config.PowerUpBarSpacing
	}

	if g.player.HasShield() {
		ui.DrawPowerUpBarAt(screen, float32(g.player.ShieldProgress()), color.RGBA{100, 200, 255, 255}, barY)
		barY += config.PowerUpBarSpacing
	}

	if g.slowMotionActive {
		ui.DrawPowerUpBarAt(screen, float32(g.slowMotionTimer.Progress()), color.RGBA{100, 255, 255, 255}, barY)
		barY += config.PowerUpBarSpacing
	}

	if g.laserBeamActive {
		ui.DrawPowerUpBarAt(screen, float32(g.laserBeamTimer.Progress()), color.RGBA{150, 100, 255, 255}, barY)
		barY += config.PowerUpBarSpacing
	}

	if g.nukeActive {
		ui.DrawPowerUpBarAt(screen, float32(g.nukeTimer.Progress()), color.RGBA{255, 50, 50, 255}, barY)
		barY += 30
	}

	if g.multiplierActive {
		ui.DrawPowerUpBarAt(screen, float32(g.multiplierTimer.Progress()), color.RGBA{255, 215, 0, 255}, barY)
	}
}

func (g *Game) drawMobileControls(screen *ebiten.Image) {
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

	for _, p := range g.powerUps {
		p.Draw(screen)
	}

	for _, c := range g.coins {
		c.Draw(screen)
	}

	g.drawParticlesBatch(screen)

	g.drawUI(screen)

	alpha := uint8(255)
	if g.bossAnnouncementTimer < config.BossAnnouncementFade {
		alpha = uint8(float64(g.bossAnnouncementTimer) / float64(config.BossAnnouncementFade) * 255)
	}

	g.drawBossAnnouncementText(screen, alpha)
}

func (g *Game) drawBossAnnouncementText(screen *ebiten.Image, alpha uint8) {
	txt := "BOSS INCOMING!"
	face := assets.FontUi
	txtWidth := measureText(txt, face)
	x := (config.ScreenWidth - txtWidth) / 2
	y := config.ScreenHeight / 2

	shadowOffset := 4
	drawText(screen, txt, face, x+shadowOffset, y+shadowOffset, color.RGBA{0, 0, 0, alpha})
	drawText(screen, txt, face, x, y, color.RGBA{255, 50, 50, alpha})
}

func (g *Game) drawBossFight(screen *ebiten.Image) {
	g.player.Draw(screen)

	if g.boss != nil {
		g.boss.Draw(screen)

		minions := g.boss.GetMinions()
		for i, minion := range minions {
			if minion != nil {
				minion.Draw(screen)
			}
			_ = i
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
	youDiedX := (config.ScreenWidth - measureText(youDiedText, assets.FontUi)) / 2
	drawText(screen, youDiedText, assets.FontUi, youDiedX, 150, color.White)

	if g.statistics != nil {
		g.statistics.Draw(screen, 220)
	}

	tryAgainText := "Press ENTER to try again"
	tryAgainX := (config.ScreenWidth - measureText(tryAgainText, assets.FontUi)) / 2
	drawText(screen, tryAgainText, assets.FontUi, tryAgainX, 480, color.White)

	g.drawShopHint(screen)

	highScoreText := fmt.Sprintf("HIGH SCORE: %d", g.highScore)
	highScoreX := config.ScreenWidth - measureText(highScoreText, assets.FontSmall) - 20
	drawText(screen, highScoreText, assets.FontSmall, highScoreX, 570, color.RGBA{255, 215, 0, 255})
}

func (g *Game) drawShopHint(screen *ebiten.Image) {
	if g.isMobile {
		g.drawShopIconButton(screen)
	} else {
		shopText := "Press S to open SHOP"
		shopX := (config.ScreenWidth - measureText(shopText, assets.FontSmall)) / 2
		drawText(screen, shopText, assets.FontSmall, shopX, 535, color.RGBA{255, 215, 0, 255})
	}
}

func (g *Game) drawShopIconButton(screen *ebiten.Image) {
	const btnX, btnY, btnSize = 10.0, 10.0, 35.0
	x, y := ebiten.CursorPosition()

	iconOp := &ebiten.DrawImageOptions{}
	if float64(x) >= btnX && float64(x) <= btnX+btnSize && float64(y) >= btnY && float64(y) <= btnY+btnSize {
		iconOp.ColorScale.ScaleWithColor(color.RGBA{255, 215, 0, 255})
	}

	scale := btnSize / float64(assets.CoinSprite.Bounds().Dx())
	iconOp.GeoM.Scale(scale, scale)
	iconOp.GeoM.Translate(btnX, btnY)
	screen.DrawImage(assets.CoinSprite, iconOp)
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
