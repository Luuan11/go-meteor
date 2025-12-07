//go:build !js || !wasm
// +build !js !wasm

package core

import "time"

func (g *Game) notifyWebLeaderboard(_ string, _ int) {
}

func (g *Game) showNameInputModal() {
}

func (g *Game) hasNameInputModal() bool {
	return false
}

func (g *Game) initNewGameSession() {
	g.meteorsDestroyed = 0
	g.powerUpsCollected = 0
	g.gameStartTime = time.Now()
	g.survivalTime = 0
}
