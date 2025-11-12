//go:build js && wasm
// +build js,wasm

package core

import (
	"go-meteor/internal/config"
	"syscall/js"
)

func (g *Game) notifyWebLeaderboard(name string, score int) {
	js.Global().Get("console").Call("log", "[GO] notifyWebLeaderboard called with name:", name, "score:", score)
	updateFunc := js.Global().Get("updateLeaderboard")
	js.Global().Get("console").Call("log", "[GO] updateLeaderboard function found:", !updateFunc.IsUndefined() && !updateFunc.IsNull())
	if !updateFunc.IsUndefined() && !updateFunc.IsNull() {
		js.Global().Get("console").Call("log", "[GO] Invoking updateLeaderboard")
		updateFunc.Invoke(name, score)
		js.Global().Get("console").Call("log", "[GO] updateLeaderboard invoked successfully")
	} else {
		js.Global().Get("console").Call("log", "[GO] ERROR: updateLeaderboard is undefined or null")
	}
}

func (g *Game) showNameInputModal() {
	showModal := js.Global().Get("showNameInputModal")
	if showModal.IsUndefined() || showModal.IsNull() {
		g.state = config.StateGameOver
		return
	}

	callback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		js.Global().Get("console").Call("log", "[GO] Modal callback executed, args length:", len(args))
		if len(args) > 0 {
			name := args[0].String()
			js.Global().Get("console").Call("log", "[GO] Name received from modal:", name)
			if name != "" {
				g.playerName = name
				g.leaderboard.AddScore(name, g.score)
				js.Global().Get("console").Call("log", "[GO] Added score to local leaderboard")

				data, err := g.leaderboard.ToJSON()
				if err == nil {
					g.storage.SaveLeaderboard(data)
					js.Global().Get("console").Call("log", "[GO] Saved to local storage")
				}

				g.notifyWebLeaderboard(name, g.score)
			} else {
				js.Global().Get("console").Call("log", "[GO] Name is empty, skipping save")
			}
		} else {
			js.Global().Get("console").Call("log", "[GO] No args received from modal")
		}
		g.state = config.StateGameOver
		js.Global().Get("console").Call("log", "[GO] State set to GameOver")
		return nil
	})

	showModal.Invoke(callback)
}

func (g *Game) hasNameInputModal() bool {
	return true
}
