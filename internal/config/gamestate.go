package config

type GameState int

const (
	StateMenu GameState = iota
	StatePlaying
	StatePaused
	StateGameOver
	StateBossFight
	StateWaitingNameInput
)

type BossType int

const (
	BossTank BossType = iota
	BossSniper
	BossSwarm
)
