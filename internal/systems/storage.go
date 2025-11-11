//go:build !js && !wasm
// +build !js,!wasm

package systems

import (
	"os"
	"path/filepath"
	"strconv"
)

type Storage interface {
	SaveHighScore(score int) error
	LoadHighScore() int
	SaveLeaderboard(data string) error
	LoadLeaderboard() (string, error)
}

type localStorage struct {
	dataDir string
}

func NewStorage() Storage {
	homeDir, _ := os.UserHomeDir()
	dataDir := filepath.Join(homeDir, ".go-meteor")
	os.MkdirAll(dataDir, 0755)
	return &localStorage{dataDir: dataDir}
}

func (s *localStorage) SaveHighScore(score int) error {
	path := filepath.Join(s.dataDir, "highscore.txt")
	return os.WriteFile(path, []byte(strconv.Itoa(score)), 0644)
}

func (s *localStorage) LoadHighScore() int {
	path := filepath.Join(s.dataDir, "highscore.txt")
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	score, _ := strconv.Atoi(string(data))
	return score
}

func (s *localStorage) SaveLeaderboard(jsonData string) error {
	path := filepath.Join(s.dataDir, "leaderboard.json")
	return os.WriteFile(path, []byte(jsonData), 0644)
}

func (s *localStorage) LoadLeaderboard() (string, error) {
	path := filepath.Join(s.dataDir, "leaderboard.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return "{\"entries\":[]}", nil
	}
	return string(data), nil
}
