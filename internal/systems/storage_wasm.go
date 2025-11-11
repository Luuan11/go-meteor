//go:build js && wasm
// +build js,wasm

package systems

import "syscall/js"

type Storage interface {
	SaveHighScore(score int) error
	LoadHighScore() int
	SaveLeaderboard(data string) error
	LoadLeaderboard() (string, error)
}

type webStorage struct {
	localStorage js.Value
}

func NewStorage() Storage {
	return &webStorage{
		localStorage: js.Global().Get("localStorage"),
	}
}

func (s *webStorage) SaveHighScore(score int) error {
	s.localStorage.Call("setItem", "spaceGoHighScore", score)
	return nil
}

func (s *webStorage) LoadHighScore() int {
	val := s.localStorage.Call("getItem", "spaceGoHighScore")
	if val.IsNull() {
		return 0
	}
	if val.Type() == js.TypeNumber {
		return val.Int()
	}
	strVal := val.String()
	if strVal == "" {
		return 0
	}
	score := 0
	for _, c := range strVal {
		if c >= '0' && c <= '9' {
			score = score*10 + int(c-'0')
		}
	}
	return score
}

func (s *webStorage) SaveLeaderboard(jsonData string) error {
	s.localStorage.Call("setItem", "spaceGoLeaderboard", jsonData)
	return nil
}

func (s *webStorage) LoadLeaderboard() (string, error) {
	val := s.localStorage.Call("getItem", "spaceGoLeaderboard")
	if val.IsNull() {
		return "{\"entries\":[]}", nil
	}
	return val.String(), nil
}
