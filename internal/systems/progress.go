package systems

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

type PlayerProgress struct {
	Coins         int            `json:"coins"`
	CoinsLifetime int            `json:"coinsLifetime"`
	Upgrades      map[string]int `json:"upgrades"`
	Version       int            `json:"version"`
	Checksum      string         `json:"checksum"`
}

const (
	ProgressVersion    = 1
	MaxUpgradeLevel    = 5
	MaxReasonableCoins = 100000
	checksumSalt       = "go-meteor-shop-v1"
)

func NewPlayerProgress() *PlayerProgress {
	return &PlayerProgress{
		Coins:         0,
		CoinsLifetime: 0,
		Upgrades:      make(map[string]int),
		Version:       ProgressVersion,
	}
}

func (p *PlayerProgress) AddCoins(amount int) {
	if amount > 0 {
		p.Coins += amount
		p.CoinsLifetime += amount
	}
}

func (p *PlayerProgress) SpendCoins(amount int) bool {
	if p.Coins >= amount {
		p.Coins -= amount
		return true
	}
	return false
}

func (p *PlayerProgress) GetUpgradeLevel(powerType string) int {
	level, exists := p.Upgrades[powerType]
	if !exists {
		return 0
	}
	return level
}

func (p *PlayerProgress) UpgradePower(powerType string) bool {
	currentLevel := p.GetUpgradeLevel(powerType)
	if currentLevel >= MaxUpgradeLevel {
		return false
	}
	p.Upgrades[powerType] = currentLevel + 1
	return true
}

func (p *PlayerProgress) CalculateChecksum() string {
	data := fmt.Sprintf("%d|%d|%v|%d|%s",
		p.Coins, p.CoinsLifetime, p.Upgrades, p.Version, checksumSalt)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func (p *PlayerProgress) UpdateChecksum() {
	p.Checksum = p.CalculateChecksum()
}

func (p *PlayerProgress) Validate() error {
	if p.Coins < 0 || p.Coins > MaxReasonableCoins {
		return fmt.Errorf("invalid coins: %d", p.Coins)
	}
	if p.CoinsLifetime < p.Coins {
		return fmt.Errorf("lifetime coins cannot be less than current coins")
	}
	if p.CoinsLifetime > MaxReasonableCoins {
		return fmt.Errorf("invalid lifetime coins: %d", p.CoinsLifetime)
	}
	for powerType, level := range p.Upgrades {
		if level < 0 || level > MaxUpgradeLevel {
			return fmt.Errorf("invalid upgrade level for %s: %d", powerType, level)
		}
	}
	expectedChecksum := p.CalculateChecksum()
	if p.Checksum != expectedChecksum {
		return fmt.Errorf("checksum mismatch: data may be corrupted")
	}
	return nil
}

func (p *PlayerProgress) ToJSON() (string, error) {
	p.UpdateChecksum()
	data, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func PlayerProgressFromJSON(jsonData string) (*PlayerProgress, error) {
	var progress PlayerProgress
	if err := json.Unmarshal([]byte(jsonData), &progress); err != nil {
		return nil, err
	}
	if progress.Upgrades == nil {
		progress.Upgrades = make(map[string]int)
	}
	if err := progress.Validate(); err != nil {
		return nil, err
	}
	return &progress, nil
}
