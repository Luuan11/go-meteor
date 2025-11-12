//go:build js && wasm
// +build js,wasm

package core

import (
	"go-meteor/internal/config"
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
		g.state = config.StateGameOver
		return
	}

	callback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			name := args[0].String()
			if name != "" {
				g.playerName = name
				g.leaderboard.AddScore(name, g.score)
				js.Global().Get("console").Call("log", "✅ Score saved to leaderboard:", name, "-", g.score, "points")

				data, err := g.leaderboard.ToJSON()
				if err == nil {
					g.storage.SaveLeaderboard(data)
					js.Global().Get("console").Call("log", "💾 Leaderboard saved to local storage")
				}

				g.notifyWebLeaderboard(name, g.score)
			}
		}
		g.state = config.StateGameOver
		return nil
	})

	showModal.Invoke(callback)
}

func (g *Game) hasNameInputModal() bool {
	return true
}
