//go:build !js || !wasm
// +build !js !wasm

package core

func (g *Game) notifyWebLeaderboard(name string, score int) {
}
