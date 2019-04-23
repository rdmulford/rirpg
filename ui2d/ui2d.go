// Riley Mulford April 2019
package ui2d

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

type ui struct {
	winWidth          int
	winHeight         int
	renderer          *sdl.Renderer
	window            *sdl.Window
	textureAtlas      *sdl.Texture
	textureIndex      map[game.Tile][]sdl.Rect
	prevKeyboardState []uint8
	keyboardState     []uint8
	centerX           int
	centerY           int
	r                 *rand.Rand
	levelChan         chan *game.Level
	inputChan         chan *game.Input
}

// init - initialize sdl
func init() {
	// Initialize SDL
	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		panic(err)
	}
}

func NewUI(inputChan chan *game.Input, levelChan chan *game.Level) *ui {
	ui := &ui{}
	ui.inputChan = inputChan
	ui.levelChan = levelChan
	ui.r = rand.New(rand.NewSource(1))
	ui.winHeight = 720
	ui.winWidth = 1280

	// Initialize window
	window, err := sdl.CreateWindow("rirpg", 200, 200, int32(ui.winWidth), int32(ui.winHeight), sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	ui.window = window

	// Initialize renderer
	ui.renderer, err = sdl.CreateRenderer(ui.window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		panic(err)
	}

	sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "1") //bilinear filtering

	ui.textureAtlas = ui.imgFileToTexture("ui2d/assets/tiles/tiles.png")
	ui.loadTextureIndex()

	ui.keyboardState = sdl.GetKeyboardState()
	ui.prevKeyboardState = make([]uint8, len(ui.keyboardState))
	for i, v := range ui.keyboardState {
		ui.prevKeyboardState[i] = v
	}

	ui.centerX = -1
	ui.centerY = -1
	return ui
}

// loadTextureIndex - Parse atlas-index.txt file to obtain coordinates for each defined tile
func (ui *ui) loadTextureIndex() {
	ui.textureIndex = make(map[game.Tile][]sdl.Rect)
	infile, err := os.Open("ui2d/assets/tiles/atlas-index.txt")
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

		ui.textureIndex[tileRune] = rects
	}
}

// imgFileToTexture - Create sdl texture from given image
func (ui *ui) imgFileToTexture(filename string) *sdl.Texture {
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
	tex, err := ui.renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, int32(w), int32(h))
	if err != nil {
		panic(err)
	}
	tex.Update(nil, pixels, w*4)

	err = tex.SetBlendMode(sdl.BLENDMODE_BLEND)
	if err != nil {
		panic(err)
	}
	return tex
}

// Draw - Given level information, draw all tiles into the window
func (ui *ui) Draw(level *game.Level) {
	// calculate scrolling
	if ui.centerX == -1 && ui.centerY == -1 {
		ui.centerX = level.Player.X
		ui.centerY = level.Player.Y
	}
	limit := 5
	if level.Player.X > ui.centerX+limit {
		ui.centerX++
	} else if level.Player.X < ui.centerX-limit {
		ui.centerX--
	} else if level.Player.Y > ui.centerY+limit {
		ui.centerY++
	} else if level.Player.Y < ui.centerY-limit {
		ui.centerY--
	}
	offsetX := int32((ui.winWidth / 2) - ui.centerX*32)
	offsetY := int32((ui.winHeight / 2) - ui.centerY*32)

	ui.renderer.Clear()
	ui.r.Seed(1)
	for y, row := range level.Map {
		for x, tile := range row {
			if tile != game.Blank {
				srcRects := ui.textureIndex[tile]
				srcRect := srcRects[ui.r.Intn(len(srcRects))]
				dstRect := sdl.Rect{int32(x*32) + offsetX, int32(y*32) + offsetY, int32(32), int32(32)}

				// debug map drawing
				pos := game.Pos{x, y}
				if level.Debug[pos] {
					ui.textureAtlas.SetColorMod(128, 0, 0)
				} else {
					// no change to texture
					ui.textureAtlas.SetColorMod(255, 255, 255)
				}

				ui.renderer.Copy(ui.textureAtlas, &srcRect, &dstRect)
			}
		}
	}
	ui.renderer.Copy(ui.textureAtlas, &sdl.Rect{int32(21 * 32), int32(59 * 32), 32, 32}, &sdl.Rect{int32(level.Player.X)*32 + offsetX, int32(level.Player.Y)*32 + offsetY, 32, 32})
	ui.renderer.Present()
}

func (ui *ui) Run() {
	for {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			// check event type and react
			switch e := event.(type) {
			case *sdl.QuitEvent:
				ui.inputChan <- &game.Input{Typ: game.QuitGame}
			case *sdl.WindowEvent:
				if e.Event == sdl.WINDOWEVENT_CLOSE {
					ui.inputChan <- &game.Input{Typ: game.CloseWindow, LevelChannel: ui.levelChan}
				}
			}
		}

		select {
		case newLevel, ok := <-ui.levelChan:
			if ok {
				ui.Draw(newLevel)
			}
		default:
		}

		if sdl.GetKeyboardFocus() == ui.window || sdl.GetMouseFocus() == ui.window {
			var input game.Input
			if ui.keyboardState[sdl.SCANCODE_UP] == 0 && ui.prevKeyboardState[sdl.SCANCODE_UP] != 0 {
				input.Typ = game.Up
			} else if ui.keyboardState[sdl.SCANCODE_DOWN] == 0 && ui.prevKeyboardState[sdl.SCANCODE_DOWN] != 0 {
				input.Typ = game.Down
			} else if ui.keyboardState[sdl.SCANCODE_LEFT] == 0 && ui.prevKeyboardState[sdl.SCANCODE_LEFT] != 0 {
				input.Typ = game.Left
			} else if ui.keyboardState[sdl.SCANCODE_RIGHT] == 0 && ui.prevKeyboardState[sdl.SCANCODE_RIGHT] != 0 {
				input.Typ = game.Right
			} else if ui.keyboardState[sdl.SCANCODE_S] == 0 && ui.prevKeyboardState[sdl.SCANCODE_S] != 0 {
				input.Typ = game.Search
			}

			for i, v := range ui.keyboardState {
				ui.prevKeyboardState[i] = v
			}

			if input.Typ != game.None {
				ui.inputChan <- &input
			}
		}
		sdl.Delay(10)
	}
}