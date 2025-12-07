package core

import "github.com/hajimehoshi/ebiten/v2"

type Updatable interface {
	Update()
}

type Drawable interface {
	Draw(screen *ebiten.Image)
}

func updateEntities[T Updatable](entities []T) {
	for _, entity := range entities {
		entity.Update()
	}
}

func (g *Game) updateAllEntities() {
	for _, p := range g.powerUps {
		p.Update()
	}
	for _, l := range g.lasers {
		l.Update()
	}
	for _, p := range g.particles {
		p.Update()
	}
	for _, c := range g.coins {
		c.Update()
	}
}

func (g *Game) updateBossEntities() {
	if g.boss == nil {
		return
	}

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

func (g *Game) detectMobileTouch(touchIDs []ebiten.TouchID) {
	if len(touchIDs) > 0 && !g.touchDetected {
		g.touchDetected = true
		g.isMobile = true
		g.initMobileControls()
	}
}
