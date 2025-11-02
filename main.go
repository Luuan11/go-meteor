package main

import (
	"go-meteor/internal/core"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	g := core.NewGame()

	err := ebiten.RunGame(g)
	if err != nil {
		log.Println("Error: running game", err)
		panic(err)
	}
}
