package worldgen

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
)

// utilize perlin noise to generate new level file at game/maps/level1.map
func GenerateNewLevel(xSize, ySize int, seed int64) {
	genMap := make([][]rune, ySize)
	for i := range genMap {
		genMap[i] = make([]rune, xSize)
	}
	p := NewPerlin(2, 2, 3, seed)

	for y := 0; y < ySize; y++ {
		for x := 0; x < xSize; x++ {
			val := p.Noise2D(float64(x)/10, float64(y)/10)
			if x == 0 || x == xSize-1 || y == 0 || y == ySize-1 {
				genMap[x][y] = '#'
				continue
			}
			if val < 0.0 {
				genMap[x][y] = '.'
			} else {
				genMap[x][y] = ','
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

	// place the player, gonna need to make more complicated later
	genMap[25][25] = '@'

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
