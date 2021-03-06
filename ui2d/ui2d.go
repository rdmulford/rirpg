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
	"github.com/veandco/go-sdl2/ttf"
)

type ui struct {
	winWidth          int
	winHeight         int
	renderer          *sdl.Renderer
	window            *sdl.Window
	textureAtlas      *sdl.Texture
	textureIndex      map[rune][]sdl.Rect // maps character from map to sprite sheet
	prevKeyboardState []uint8
	keyboardState     []uint8
	centerX           int
	centerY           int
	r                 *rand.Rand
	levelChan         chan *game.Level
	inputChan         chan *game.Input
	fontSmall         *ttf.Font
	fontMedium        *ttf.Font
	fontLarge         *ttf.Font
	str2TexSmall      map[string]*sdl.Texture
	str2TexMedium     map[string]*sdl.Texture
	str2TexLarge      map[string]*sdl.Texture
	eventBackground   *sdl.Texture
}

// init - initialize sdl
func init() {
	// Initialize SDL
	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		panic(err)
	}

	err = ttf.Init()
	if err != nil {
		panic(err)
	}
}

func NewUI(inputChan chan *game.Input, levelChan chan *game.Level) *ui {
	ui := &ui{}
	ui.inputChan = inputChan
	ui.levelChan = levelChan
	ui.r = rand.New(rand.NewSource(1))
	ui.winHeight = 1080
	ui.winWidth = 1920
	ui.str2TexSmall = make(map[string]*sdl.Texture)
	ui.str2TexMedium = make(map[string]*sdl.Texture)
	ui.str2TexLarge = make(map[string]*sdl.Texture)

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

	// Set up fonts
	ui.fontSmall, err = ttf.OpenFont("ui2d/assets/fonts/gothic.ttf", int(float64(ui.winWidth)*0.015))
	if err != nil {
		panic(err)
	}
	ui.fontMedium, err = ttf.OpenFont("ui2d/assets/fonts/gothic.ttf", 24)
	if err != nil {
		panic(err)
	}
	ui.fontLarge, err = ttf.OpenFont("ui2d/assets/fonts/gothic.ttf", 32)
	if err != nil {
		panic(err)
	}

	ui.eventBackground = ui.GetSinglePixelTex(sdl.Color{0, 0, 0, 128})
	ui.eventBackground.SetBlendMode(sdl.BLENDMODE_BLEND)

	return ui
}

type FontSize int

const (
	FontSmall FontSize = iota
	FontMedium
	FontLarge
)

// TODO remove textures from string to texture caches
// this funciton is really expensive, call as little times as possible
func (ui *ui) stringToTexture(s string, color sdl.Color, size FontSize) *sdl.Texture {
	var font *ttf.Font
	switch size {
	case FontSmall:
		font = ui.fontSmall
		tex, exists := ui.str2TexSmall[s]
		if exists {
			return tex
		}
	case FontMedium:
		font = ui.fontMedium
		tex, exists := ui.str2TexMedium[s]
		if exists {
			return tex
		}
	case FontLarge:
		font = ui.fontLarge
		tex, exists := ui.str2TexLarge[s]
		if exists {
			return tex
		}
	}
	fontSurface, err := font.RenderUTF8Blended(s, color)
	if err != nil {
		panic(err)
	}
	tex, err := ui.renderer.CreateTextureFromSurface(fontSurface)
	if err != nil {
		panic(err)
	}

	switch size {
	case FontSmall:
		ui.str2TexSmall[s] = tex
	case FontMedium:
		ui.str2TexMedium[s] = tex
	case FontLarge:
		ui.str2TexLarge[s] = tex
	}

	return tex
}

// loadTextureIndex - Parse atlas-index.txt file to obtain coordinates for each defined tile
func (ui *ui) loadTextureIndex() {
	ui.textureIndex = make(map[rune][]sdl.Rect)
	infile, err := os.Open("ui2d/assets/tiles/atlas-index.txt")
	if err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(infile)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		tileRune := rune(line[0])
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
		for i := int64(0); i < variationCount; i++ {
			rects = append(rects, sdl.Rect{int32(x * 32), int32(y * 32), int32(32), int32(32)})
			// atlas wraps around
			x++
			if x > 62 {
				x = 0
				y++
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
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			pixels[bIndex] = byte(r / 256)
			bIndex++
			pixels[bIndex] = byte(g / 256)
			bIndex++
			pixels[bIndex] = byte(b / 256)
			bIndex++
			pixels[bIndex] = byte(a / 256)
			bIndex++
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
	// draws all floor tiles
	for y, row := range level.Map {
		for x, tile := range row {
			// draw grass under trees
			// TODO refactor how this works in general
			drawnTile := tile.Symbol
			if tile.Symbol == game.Tree {
				drawnTile = game.Grass
			}
			if tile.Symbol != game.Blank {
				srcRects := ui.textureIndex[drawnTile]
				srcRect := srcRects[ui.r.Intn(len(srcRects))]
				if tile.Visible || tile.Seen {
					dstRect := sdl.Rect{int32(x*32) + offsetX, int32(y*32) + offsetY, int32(32), int32(32)}

					// debug map drawing
					pos := game.Pos{x, y}
					if level.Debug[pos] {
						ui.textureAtlas.SetColorMod(128, 0, 0)
					} else if tile.Seen && !tile.Visible {
						ui.textureAtlas.SetColorMod(128, 128, 128)
					} else {
						ui.textureAtlas.SetColorMod(255, 255, 255)
					}

					ui.renderer.Copy(ui.textureAtlas, &srcRect, &dstRect)
				}
			}
		}
	}

	// TODO clean up this logic
	ui.textureAtlas.SetColorMod(255, 255, 255)
	// draw trees
	for pos, tree := range level.Trees {
		if level.Map[pos.Y][pos.X].Seen && !level.Map[pos.Y][pos.X].Visible {
			ui.textureAtlas.SetColorMod(128, 128, 128)
		}
		if level.Map[pos.Y][pos.X].Visible || level.Map[pos.Y][pos.X].Seen {
			treeSrcRect := ui.textureIndex[tree.Symbol][0]
			ui.renderer.Copy(ui.textureAtlas, &treeSrcRect, &sdl.Rect{int32(pos.X)*32 + offsetX, int32(pos.Y)*32 + offsetY, 32, 32})
		}
		ui.textureAtlas.SetColorMod(255, 255, 255)
	}

	// draw monsters
	for pos, monster := range level.Monsters {
		if level.Map[pos.Y][pos.X].Visible {
			monsterSrcRect := ui.textureIndex[monster.Symbol][0]
			ui.renderer.Copy(ui.textureAtlas, &monsterSrcRect, &sdl.Rect{int32(pos.X)*32 + offsetX, int32(pos.Y)*32 + offsetY, 32, 32})
		}
	}

	// draws player
	playerSrcRect := ui.textureIndex['@'][0]
	ui.renderer.Copy(ui.textureAtlas, &playerSrcRect, &sdl.Rect{int32(level.Player.X)*32 + offsetX, int32(level.Player.Y)*32 + offsetY, 32, 32})

	// draw text events
	// TODO scroll better
	textStart := int32(float64(ui.winHeight) * 0.68)
	textWidth := int32(float64(ui.winWidth) * 0.25)
	// draw text event background
	ui.renderer.Copy(ui.eventBackground, nil, &sdl.Rect{0, textStart, textWidth, int32(ui.winHeight) - textStart})
	i := level.EventPos
	count := 0
	_, fontSizeY, _ := ui.fontSmall.SizeUTF8("A")
	for {
		event := level.Events[i]
		if event != "" {
			tex := ui.stringToTexture(event, sdl.Color{255, 0, 0, 0}, FontSmall)
			_, _, w, h, err := tex.Query()
			if err != nil {
				panic(err)
			}
			ui.renderer.Copy(tex, nil, &sdl.Rect{5, int32(count*fontSizeY) + textStart, w, h})
		}
		i = (i + 1) % len(level.Events)
		count++
		if i == level.EventPos {
			break
		}
	}

	ui.renderer.Present()
}

// key pressed
func (ui *ui) keyDownOnce(key uint8) bool {
	return ui.keyboardState[key] == 1 && ui.prevKeyboardState[key] == 0
}

// key press then released
func (ui *ui) keyPressed(key uint8) bool {
	return ui.keyboardState[key] == 0 && ui.prevKeyboardState[key] == 1
}

// utility function
func (ui *ui) GetSinglePixelTex(color sdl.Color) *sdl.Texture {
	tex, err := ui.renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STATIC, 1, 1)
	if err != nil {
		panic(err)
	}
	pixels := make([]byte, 4)
	pixels[0] = color.R
	pixels[1] = color.G
	pixels[2] = color.B
	pixels[3] = color.A
	tex.Update(nil, pixels, 4)
	return tex
}

func (ui *ui) Run() {
	for {
		// TODO if we want multiple uis, need to seperate this into diffrent component on main thread
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

		// TODO suspect quick keypress cause channel deadlock
		select {
		case newLevel, ok := <-ui.levelChan:
			if ok {
				ui.Draw(newLevel)
			}
		default:
		}

		// TODO made a function to ask "has a key been pressed"
		if sdl.GetKeyboardFocus() == ui.window || sdl.GetMouseFocus() == ui.window {
			var input game.Input
			if ui.keyDownOnce(sdl.SCANCODE_UP) {
				input.Typ = game.Up
			} else if ui.keyDownOnce(sdl.SCANCODE_DOWN) {
				input.Typ = game.Down
			} else if ui.keyDownOnce(sdl.SCANCODE_LEFT) {
				input.Typ = game.Left
			} else if ui.keyDownOnce(sdl.SCANCODE_RIGHT) {
				input.Typ = game.Right
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
