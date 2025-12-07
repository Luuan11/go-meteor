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

const GameVersion = "0.2.0"

func getSecretKey() []byte {
	encrypted := []byte{
		0x67, 0x6f, 0x2d, 0x6d, 0x65, 0x74, 0x65, 0x6f, 0x72, 0x2d,
		0x73, 0x65, 0x63, 0x72, 0x65, 0x74, 0x2d, 0x32, 0x30, 0x32,
		0x35, 0x2d, 0x73, 0x65, 0x63, 0x75, 0x72, 0x65,
	}
	key := make([]byte, len(encrypted))
	for i, b := range encrypted {
		key[i] = b ^ 0xAA
	}
	return key
}

func generateSignature(name string, score int, sessionToken string, timestamp int64) string {
	message := fmt.Sprintf("%s|%d|%s|%d", name, score, sessionToken, timestamp)
	h := hmac.New(sha256.New, getSecretKey())
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}

func (g *Game) notifyWebLeaderboard(name string, score int) {
	js.Global().Get("console").Call("log", "[Game Version]", GameVersion)

	updateFunc := js.Global().Get("updateLeaderboard")
	if updateFunc.IsUndefined() || updateFunc.IsNull() {
		js.Global().Get("console").Call("error", "[Security] updateLeaderboard function not found")
		return
	}

	sessionTokenValue := js.Global().Get("gameSessionToken")
	if sessionTokenValue.IsUndefined() || sessionTokenValue.IsNull() || sessionTokenValue.String() == "" {
		js.Global().Get("console").Call("error", "[Security] No session token available")
		return
	}

	sessionToken := sessionTokenValue.String()
	timestamp := time.Now().UnixMilli()
	signature := generateSignature(name, score, sessionToken, timestamp)

	js.Global().Get("console").Call("log", "[Security] Sending score with HMAC signature")
	updateFunc.Invoke(name, score, signature, timestamp)
}

func (g *Game) showNameInputModal() {
	isTopScore := js.Global().Get("isTopScore")
	if !isTopScore.IsUndefined() && !isTopScore.IsNull() {
		promiseCallback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			defer func() {
				if r := recover(); r != nil {
					js.Global().Get("console").Call("error", "[Promise] Panic in callback:", r)
				}
			}()

			if len(args) > 0 {
				isTop := args[0].Bool()
				if !isTop {
					g.state = config.StateGameOver
					return nil
				}

				g.showModalInternal()
			}
			return nil
		})

		promise := isTopScore.Invoke(g.score)
		promise.Call("then", promiseCallback)
	} else {
		g.showModalInternal()
	}
}

func (g *Game) showModalInternal() {
	showModal := js.Global().Get("showNameInputModal")
	if showModal.IsUndefined() || showModal.IsNull() {
		js.Global().Get("console").Call("error", "[Modal] showNameInputModal function not found")
		g.state = config.StateGameOver
		return
	}

	var callback js.Func
	callback = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		defer callback.Release()
		defer func() {
			if r := recover(); r != nil {
				js.Global().Get("console").Call("error", "[Modal] Panic:", r)
				g.state = config.StateGameOver
			}
		}()

		if len(args) > 0 && args[0].String() != "" {
			name := args[0].String()
			g.leaderboard.AddScore(name, g.score)

			data, err := g.leaderboard.ToJSON()
			if err == nil {
				g.storage.SaveLeaderboard(data)
			}

			g.notifyWebLeaderboard(name, g.score)
		}

		g.state = config.StateGameOver
		return nil
	})

	showModal.Invoke(callback)
}

func (g *Game) hasNameInputModal() bool {
	return true
}

func (g *Game) initNewGameSession() {
	g.meteorsDestroyed = 0
	g.powerUpsCollected = 0
	g.gameStartTime = time.Now()
	g.survivalTime = 0

	initFunc := js.Global().Get("initGameSession")
	if !initFunc.IsUndefined() && !initFunc.IsNull() {
		initFunc.Invoke()
		js.Global().Get("console").Call("log", "[Game] New session token requested")
	}
}
