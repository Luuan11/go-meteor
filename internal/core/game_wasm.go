//go:build js && wasm
// +build js,wasm

package core

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"go-meteor/internal/config"
	"syscall/js"
	"time"
)

const secretKey = "go-meteor-secret-2025-secure-key-change-in-production"

func generateSignature(name string, score int, sessionToken string, timestamp int64) string {
	message := fmt.Sprintf("%s|%d|%s|%d", name, score, sessionToken, timestamp)
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}

func (g *Game) notifyWebLeaderboard(name string, score int) {
	updateFunc := js.Global().Get("updateLeaderboard")
	if !updateFunc.IsUndefined() && !updateFunc.IsNull() {
		sessionToken := js.Global().Get("gameSessionToken")
		if sessionToken.IsUndefined() || sessionToken.IsNull() {
			js.Global().Get("console").Call("error", "[Security] No session token")
			return
		}

		timestamp := time.Now().UnixMilli()
		signature := generateSignature(name, score, sessionToken.String(), timestamp)

		updateFunc.Invoke(name, score, signature, timestamp)
	}
}

func (g *Game) showNameInputModal() {
	isTopScore := js.Global().Get("isTopScore")
	if !isTopScore.IsUndefined() && !isTopScore.IsNull() {
		promiseCallback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			if len(args) > 0 {
				isTop := args[0].Bool()
				if !isTop {
					js.Global().Get("console").Call("log", "[Leaderboard] Score not high enough for top 10")
					g.state = config.StateGameOver
					return nil
				}

				g.showModalInternal()
			}
			return nil
		})
		defer promiseCallback.Release()

		promise := isTopScore.Invoke(g.score)
		promise.Call("then", promiseCallback)
	} else {
		g.showModalInternal()
	}
}

func (g *Game) showModalInternal() {
	showModal := js.Global().Get("showNameInputModal")
	if showModal.IsUndefined() || showModal.IsNull() {
		g.state = config.StateGameOver
		return
	}

	callback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			name := args[0].String()
			if name != "" {
				g.leaderboard.AddScore(name, g.score)
				js.Global().Get("console").Call("log", "[Leaderboard] Score saved:", name, "-", g.score, "points")

				data, err := g.leaderboard.ToJSON()
				if err == nil {
					g.storage.SaveLeaderboard(data)
					js.Global().Get("console").Call("log", "[Storage] Leaderboard saved to local storage")
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
