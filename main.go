// Riley Mulford April 2019
package main

import (
	"github.com/rdmulford/rirpg/game"
	"github.com/rdmulford/rirpg/ui2d"
)

func main() {
	//game := game.NewGame(1, "game/maps/level1.map")
	game := game.NewGame(1, "")
	go func() { game.Run() }()
	ui := ui2d.NewUI(game.InputChan, game.LevelChans[0])
	ui.Run()
}
