package worldgen

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
)

type pos struct {
	x, y int
}

// utilize perlin noise to generate new level file at game/maps/level1.map
func GenerateNewLevel(xSize, ySize int, seed int64) {
	genMap := make([][]rune, ySize)
	for i := range genMap {
		genMap[i] = make([]rune, xSize)
	}
	dirtTiles := make([]pos, 0)

	p := NewPerlin(2, 2, 3, seed)
	for y := 0; y < ySize; y++ {
		for x := 0; x < xSize; x++ {
			val := p.Noise2D(float64(x)/10, float64(y)/10)
			if x == 0 || x == xSize-1 || y == 0 || y == ySize-1 {
				genMap[x][y] = '#'
				continue
			}
			if val < -0.3 {
				genMap[x][y] = '~'
			} else if val >= -0.3 && val < 0.3 {
				genMap[x][y] = ','
			} else if val >= 0.3 {
				genMap[x][y] = '.'
				dirtTiles = append(dirtTiles, pos{x, y})
			}
		}
	}

	// place monsters, overwriting old tiles (that could be bad?)
	r := rand.New(rand.NewSource(1))
	r.Seed(seed)
	for i := 0; i < 5; i++ {
		genMap[r.Intn(xSize-2)][r.Intn(ySize-2)] = 'R'
	}
	for i := 0; i < 5; i++ {
		genMap[r.Intn(xSize-2)][r.Intn(ySize-2)] = 'S'
	}

	// place trees
	for i := 0; i < 200; i++ {
		rX := r.Intn(xSize - 2)
		rY := r.Intn(ySize - 2)
		if genMap[rX][rY] == ',' {
			genMap[rX][rY] = '^'
		}
	}

	// place player
	playerPos := dirtTiles[rand.Intn(len(dirtTiles))]
	genMap[playerPos.x][playerPos.y] = '@'

	err := os.Remove("game/maps/level1.map")
	if err != nil {
		panic(err)
	}

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
