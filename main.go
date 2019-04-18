package main

import (
	"github.com/rdmulford/rirpg/game"
	"github.com/rdmulford/rirpg/ui"
)

func main() {
	gameUI := &ui.UI{}
	game.Run(gameUI)
}
