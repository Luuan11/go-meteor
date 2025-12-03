package main

import (
	"go-meteor/internal/config"
	"go-meteor/internal/core"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	g := core.NewGame()

	ebiten.SetWindowSize(config.ScreenWidth, config.ScreenHeight)
	ebiten.SetWindowTitle("Meteor Game")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetVsyncEnabled(true)
	ebiten.SetTPS(60)

	err := ebiten.RunGame(g)
	if err != nil {
		log.Println("Error: running game", err)
		panic(err)
	}
}
