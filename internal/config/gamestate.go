package config

type GameState int

const (
	StateMenu GameState = iota
	StatePlaying
	StatePaused
	StateGameOver
	StateWaitingNameInput // Aguardando input de nome no modal
)
