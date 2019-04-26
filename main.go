// Riley Mulford April 2019
package main

import (
	"math/rand"
	"time"

	"github.com/rdmulford/rirpg/game"
	"github.com/rdmulford/rirpg/ui2d"
	"github.com/rdmulford/rirpg/worldgen"
)

func main() {
	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)
	worldgen.GenerateNewLevel(50, 50, int64(r.Intn(1000)))
	game := game.NewGame(1, "game/maps/level1.map")
	go func() { game.Run() }()
	ui := ui2d.NewUI(game.InputChan, game.LevelChans[0])
	ui.Run()
}
