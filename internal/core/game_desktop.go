//go:build !js || !wasm
// +build !js !wasm

package core

func (g *Game) notifyWebLeaderboard(name string, score int) {
}

func (g *Game) showNameInputModal() {
	// Desktop não tem modal, não faz nada
	// O estado já foi setado para StateGameOver em checkCollisions
}

func (g *Game) hasNameInputModal() bool {
	return false
}
