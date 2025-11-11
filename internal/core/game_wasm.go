//go:build js && wasm
// +build js,wasm

package core

import (
	"syscall/js"
)

func (g *Game) notifyWebLeaderboard(name string, score int) {
	updateFunc := js.Global().Get("updateLeaderboard")
	if !updateFunc.IsUndefined() && !updateFunc.IsNull() {
		updateFunc.Invoke(name, score)
	}
}
