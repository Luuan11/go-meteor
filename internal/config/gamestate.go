package config

type GameState int

const (
	StateMenu GameState = iota
	StatePlaying
	StatePaused
	StateGameOver
	StateShop
	StateSettings
	StateBossAnnouncement
	StateBossFight
	StatePlayerDeath
	StateWaitingNameInput
)

type BossType int

const (
	BossTank BossType = iota
	BossSniper
	BossSwarm
)
