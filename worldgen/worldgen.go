package worldgen

import (
	"bufio"
	"fmt"
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
			if x == 25 && y == 25 {
				genMap[x][y] = 'P'
				continue
			}
			if val < 0.0 {
				genMap[x][y] = '.'
			} else {
				genMap[x][y] = ','
			}
		}
	}

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
