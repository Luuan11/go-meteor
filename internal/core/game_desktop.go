//go:build !js || !wasm
// +build !js !wasm

package core

func (g *Game) notifyWebLeaderboard(_ string, _ int) {
}

func (g *Game) showNameInputModal() {
}

func (g *Game) hasNameInputModal() bool {
	return false
}

func (g *Game) initNewGameSession() {
}
