package worldgen

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"

	"github.com/rdmulford/rirpg/game"
)

// utilize perlin noise to generate new level file at game/maps/level1.map
func GenerateNewLevel(xSize, ySize int, seed int64) {
	genMap := make([][]game.Tile, ySize)
	for i := range genMap {
		genMap[i] = make([]game.Tile, xSize)
	}
	openTiles := make([]game.Pos, 0)

	// define world tiles base don perlin noise
	p := NewPerlin(2, 2, 3, seed)
	for y := 0; y < ySize; y++ {
		for x := 0; x < xSize; x++ {
			val := p.Noise2D(float64(x)/10, float64(y)/10)
			if x == 0 || x == xSize-1 || y == 0 || y == ySize-1 {
				genMap[x][y] = game.Tile('#')
				continue
			}
			if val < -0.4 {
				genMap[x][y] = game.Tile('~')
			} else if val >= -0.4 && val < -0.3 {
				genMap[x][y] = game.Tile('$')
			} else if val >= -0.3 && val < 0.3 {
				genMap[x][y] = game.Tile(',')
				openTiles = append(openTiles, game.Pos{x, y})
			} else if val >= 0.3 {
				genMap[x][y] = game.Tile('.')
				openTiles = append(openTiles, game.Pos{x, y})
			}
		}
	}

	// place trees
	for i := 0; i < 200; i++ {
		placeTile(openTiles, genMap, game.Tile('^'))
	}

	// place monsters
	for i := 0; i < 5; i++ {
		placeTile(openTiles, genMap, game.Tile('R'))
	}
	for i := 0; i < 5; i++ {
		placeTile(openTiles, genMap, game.Tile('S'))
	}

	// place player
	placeTile(openTiles, genMap, game.Tile('@'))

	// delete old level
	err := os.Remove("game/maps/level1.map")
	if err != nil {
		panic(err)
	}

	// write genMap to new level file
	f, err := os.Create("game/maps/level1.map")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	for y := 0; y < ySize; y++ {
		for x := 0; x < xSize; x++ {
			_, err = fmt.Fprintf(w, "%v", string(genMap[x][y]))
			if err != nil {
				panic(err)
			}
		}
		_, err = fmt.Fprintf(w, "\n")
		if err != nil {
			panic(err)
		}
	}
	w.Flush()
}

// place a new tile on an open tile
func placeTile(openTiles []game.Pos, genMap [][]game.Tile, tile game.Tile) {
	index := rand.Intn(len(openTiles))
	mPos := openTiles[index]
	remove(openTiles, index)
	genMap[mPos.X][mPos.Y] = tile
}

func remove(s []game.Pos, i int) []game.Pos {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}
