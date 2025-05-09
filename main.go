package main

import (
	"go-meteor/src/application"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	g := game.NewGame()

	err := ebiten.RunGame(g)
	if err != nil {
		log.Println("Error: running game", err)
		panic(err)
	}
}
