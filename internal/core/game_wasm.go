// go:build js && wasm
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

func (g *Game) showNameInputModal() {
	showModal := js.Global().Get("showNameInputModal")
	if showModal.IsUndefined() || showModal.IsNull() {
		return
	}

	callback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			name := args[0].String()
			g.playerName = name
			g.leaderboard.AddScore(name, g.score)
			
			data, err := g.leaderboard.ToJSON()
			if err == nil {
				g.storage.SaveLeaderboard(data)
			}
			
			g.notifyWebLeaderboard(name, g.score)
		}
		g.Reset()
		return nil
	})

	showModal.Invoke(callback)
}
