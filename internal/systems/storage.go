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
	SaveProgress(progress *PlayerProgress) error
	LoadProgress() (*PlayerProgress, error)
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

func (s *localStorage) SaveProgress(progress *PlayerProgress) error {
	jsonData, err := progress.ToJSON()
	if err != nil {
		return err
	}
	path := filepath.Join(s.dataDir, "progress.json")
	backupPath := filepath.Join(s.dataDir, "progress.backup.json")
	if _, err := os.Stat(path); err == nil {
		data, readErr := os.ReadFile(path)
		if readErr == nil {
			os.WriteFile(backupPath, data, 0644)
		}
	}
	return os.WriteFile(path, []byte(jsonData), 0644)
}

func (s *localStorage) LoadProgress() (*PlayerProgress, error) {
	path := filepath.Join(s.dataDir, "progress.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return NewPlayerProgress(), nil
	}
	progress, parseErr := PlayerProgressFromJSON(string(data))
	if parseErr != nil {
		backupPath := filepath.Join(s.dataDir, "progress.backup.json")
		backupData, backupErr := os.ReadFile(backupPath)
		if backupErr == nil {
			progress, parseErr = PlayerProgressFromJSON(string(backupData))
			if parseErr == nil {
				return progress, nil
			}
		}
		return NewPlayerProgress(), nil
	}
	return progress, nil
}
