// Riley Mulford April 2019
package ui

import (
	"bufio"
	"image/png"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"github.com/rdmulford/rirpg/game"
	"github.com/veandco/go-sdl2/sdl"
)

const (
	winWidth, winHeight = 1280, 720
)

var renderer *sdl.Renderer
var textureAtlas *sdl.Texture
var textureIndex map[game.Tile][]sdl.Rect

var prevKeyboardState []uint8
var keyboardState []uint8

var centerX int
var centerY int

// Parse atlas-index.txt file to obtain coordinates for each defined tile
func loadTextureIndex() {
	textureIndex = make(map[game.Tile][]sdl.Rect)
	infile, err := os.Open("ui/assets/tiles/atlas-index.txt")
	if err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(infile)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		tileRune := game.Tile(line[0])
		xy := line[1:]
		splitXYC := strings.Split(xy, ",")
		x, err := strconv.ParseInt(strings.TrimSpace(splitXYC[0]), 10, 64)
		if err != nil {
			panic(err)
		}
		y, err := strconv.ParseInt(strings.TrimSpace(splitXYC[1]), 10, 64)
		if err != nil {
			panic(err)
		}

		// Account for n number of variations of each tile in order to randomly use variations
		variationCount, err := strconv.ParseInt(strings.TrimSpace(splitXYC[2]), 10, 64)
		if err != nil {
			panic(err)
		}
		var rects []sdl.Rect
		for i := int64(0); i < variationCount; i += 1 {
			rects = append(rects, sdl.Rect{int32(x * 32), int32(y * 32), int32(32), int32(32)})
			// atlas wraps around
			x += 1
			if x > 62 {
				x = 0
				y += 1
			}
		}

		textureIndex[tileRune] = rects
	}
}

// Create sdl texture from given image
func imgFileToTexture(filename string) *sdl.Texture {
	infile, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer infile.Close()

	// TODO switch to sdl png decoder ?
	img, err := png.Decode(infile)
	if err != nil {
		panic(err)
	}

	w := img.Bounds().Max.X
	h := img.Bounds().Max.Y

	pixels := make([]byte, w*h*4)
	bIndex := 0
	for y := 0; y < h; y += 1 {
		for x := 0; x < w; x += 1 {
			r, g, b, a := img.At(x, y).RGBA()
			pixels[bIndex] = byte(r / 256)
			bIndex += 1
			pixels[bIndex] = byte(g / 256)
			bIndex += 1
			pixels[bIndex] = byte(b / 256)
			bIndex += 1
			pixels[bIndex] = byte(a / 256)
			bIndex += 1
		}
	}
	tex := pixelsToTexture(renderer, pixels, w, h)
	err = tex.SetBlendMode(sdl.BLENDMODE_BLEND)
	if err != nil {
		panic(err)
	}
	return tex
}

// Helper function to imgFiletoTexture
func pixelsToTexture(renderer *sdl.Renderer, pixels []byte, w, h int) *sdl.Texture {
	tex, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, int32(w), int32(h))
	if err != nil {
		panic(err)
	}
	tex.Update(nil, pixels, w*4)
	return tex
}

func init() {
	sdl.LogSetAllPriority(sdl.LOG_PRIORITY_VERBOSE)

	// Initialize SDL
	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		panic(err)
	}

	// Initialize window
	window, err := sdl.CreateWindow("rirpg", 200, 200, int32(winWidth), int32(winHeight), sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}

	// Initialize renderer
	renderer, err = sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		panic(err)
	}

	sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "1") //bilinear filtering

	textureAtlas = imgFileToTexture("ui/assets/tiles/tiles.png")
	loadTextureIndex()

	keyboardState = sdl.GetKeyboardState()
	prevKeyboardState = make([]uint8, len(keyboardState))
	for i, v := range keyboardState {
		prevKeyboardState[i] = v
	}

	centerX = -1
	centerY = -1
}

type UI struct {
}

func (ui *UI) Draw(level *game.Level) {
	// calculate scrolling
	if centerX == -1 && centerY == -1 {
		centerX = level.Player.X
		centerY = level.Player.Y
	}
	limit := 5
	if level.Player.X > centerX+limit {
		centerX++
	} else if level.Player.X < centerX-limit {
		centerX--
	} else if level.Player.Y > centerY+limit {
		centerY++
	} else if level.Player.Y < centerY-limit {
		centerY--
	}
	offsetX := int32((winWidth / 2) - centerX*32)
	offsetY := int32((winHeight / 2) - centerY*32)

	renderer.Clear()
	rand.Seed(1)
	for y, row := range level.Map {
		for x, tile := range row {
			if tile != game.Blank {
				srcRects := textureIndex[tile]
				srcRect := srcRects[rand.Intn(len(srcRects))]
				dstRect := sdl.Rect{int32(x*32) + offsetX, int32(y*32) + offsetY, int32(32), int32(32)}
				renderer.Copy(textureAtlas, &srcRect, &dstRect)
			}
		}
	}
	renderer.Copy(textureAtlas, &sdl.Rect{int32(21 * 32), int32(59 * 32), 32, 32}, &sdl.Rect{int32(level.Player.X)*32 + offsetX, int32(level.Player.Y)*32 + offsetY, 32, 32})
	renderer.Present()
}

func (ui *UI) GetInput() *game.Input {
	for {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			// check event type and react
			switch event.(type) {
			case *sdl.QuitEvent:
				return &game.Input{Typ: game.Quit}
			}
		}

		var input game.Input
		if keyboardState[sdl.SCANCODE_UP] == 0 && prevKeyboardState[sdl.SCANCODE_UP] != 0 {
			input.Typ = game.Up
		} else if keyboardState[sdl.SCANCODE_DOWN] == 0 && prevKeyboardState[sdl.SCANCODE_DOWN] != 0 {
			input.Typ = game.Down
		} else if keyboardState[sdl.SCANCODE_LEFT] == 0 && prevKeyboardState[sdl.SCANCODE_LEFT] != 0 {
			input.Typ = game.Left
		} else if keyboardState[sdl.SCANCODE_RIGHT] == 0 && prevKeyboardState[sdl.SCANCODE_RIGHT] != 0 {
			input.Typ = game.Right
		}

		for i, v := range keyboardState {
			prevKeyboardState[i] = v
		}

		if input.Typ != game.None {
			return &input
		}
	}
}
