// Riley Mulford April 2019
package game

import (
	"bufio"
	"fmt"
	"math"
	"os"

	"github.com/veandco/go-sdl2/sdl"
)

type Game struct {
	LevelChans []chan *Level // send level state to multiple ui
	InputChan  chan *Input   // recieve input from multiple ui
	Level      *Level
}

func NewGame(numWindows int, levelPath string) *Game {
	levelChans := make([]chan *Level, numWindows)
	for i := range levelChans {
		levelChans[i] = make(chan *Level)
	}
	inputChan := make(chan *Input)
	return &Game{levelChans, inputChan, loadLevelFromFile(levelPath)}
}

const (
	None InputType = iota
	Up
	Down
	Left
	Right
	QuitGame
	CloseWindow
	Search // TODO remove
)

type InputType int

type Input struct {
	Typ          InputType
	LevelChannel chan *Level
}

type Tile struct {
	Symbol  rune
	Visible bool
	//visited bool
}

const (
	StoneWall  rune = '#'
	DirtFloor  rune = '.'
	Grass      rune = ','
	ClosedDoor rune = '|'
	OpenDoor   rune = '/'
	Blank      rune = 0
	Tree       rune = '^'
	Water      rune = '~'
	Sand       rune = '$'
	Pending    rune = -1
)

type Pos struct {
	X, Y int
}

type Entity struct {
	Pos
	Name   string
	Symbol rune
}

type Character struct {
	Entity
	Hitpoints     int
	Strength      int
	Speed         float64
	MaxBreath     int
	CurrentBreath int
	ActionPoints  float64
	SightRange    int
}

type Player struct {
	Character
}

type Level struct {
	Map      [][]Tile
	Player   *Player
	Monsters map[Pos]*Monster
	Trees    map[Pos]Tile
	Events   []string // TODO pull event into own struct
	EventPos int
	Debug    map[Pos]bool
}

func (level *Level) Attack(c1, c2 *Character) {
	c1.ActionPoints -= 1
	c2.Hitpoints -= c1.Strength
	level.AddEvent(fmt.Sprintf("%s(%d) Attacks %s(%d)", c1.Name, c1.Hitpoints, c2.Name, c2.Hitpoints))
}

func (level *Level) AddEvent(event string) {
	// circular array
	level.Events[level.EventPos] = event
	level.EventPos++
	if level.EventPos == len(level.Events) {
		level.EventPos = 0
	}
}

// iterate over square the size of player sight range
func (level *Level) lineOfSight() {
	pos := level.Player.Pos
	dist := level.Player.SightRange
	for y := pos.Y - dist; y <= pos.Y+dist; y++ {
		for x := pos.X - dist; x <= pos.X+dist; x++ {
			xDelta := pos.X - x
			yDelta := pos.Y - y
			d := math.Sqrt(float64(xDelta*xDelta + yDelta*yDelta))
			if d <= float64(dist) {
				level.bresenhamVisibility(pos, Pos{x, y})
			}
		}
	}
}

// bresenham adapted specifically to calculate FOW
func (level *Level) bresenhamVisibility(start Pos, end Pos) {
	steep := math.Abs(float64(end.Y-start.Y)) > math.Abs(float64(end.X-start.X))
	if steep {
		start.X, start.Y = start.Y, start.X
		end.X, end.Y = end.Y, end.X
	}
	deltaY := int(math.Abs(float64(end.Y - start.Y)))
	err := 0
	y := start.Y
	ystep := 1
	if start.Y >= end.Y {
		ystep = -1
	}
	if start.X > end.X {
		deltaX := start.X - end.X
		for x := start.X; x > end.X; x-- {
			var pos Pos
			if steep {
				pos = Pos{y, x}
			} else {
				pos = Pos{x, y}
			}
			level.Map[pos.Y][pos.X].Visible = true
			if !canSeeThrough(level, pos) {
				return
			}
			err += deltaY
			if 2*err >= deltaX {
				y += ystep
				err -= deltaX
			}
		}
	} else {
		deltaX := end.X - start.X
		for x := start.X; x < end.X; x++ {
			var pos Pos
			if steep {
				pos = Pos{y, x}
			} else {
				pos = Pos{x, y}
			}
			level.Map[pos.Y][pos.X].Visible = true
			if !canSeeThrough(level, pos) {
				return
			}
			err += deltaY
			if 2*err >= deltaX {
				y += ystep
				err -= deltaX
			}
		}
	}
}

// general purpose bresenham function (might delete later idk)
func bresenham(start Pos, end Pos) []Pos {
	result := make([]Pos, 0)
	steep := math.Abs(float64(end.Y-start.Y)) > math.Abs(float64(end.X-start.X))
	if steep {
		start.X, start.Y = start.Y, start.X
		end.X, end.Y = end.Y, end.X
	}
	if start.X > end.X {
		start.X, end.X = end.X, start.X
		start.Y, end.Y = end.Y, start.Y
	}

	deltaX := end.X - start.X
	deltaY := int(math.Abs(float64(end.Y - start.Y)))
	err := 0
	y := start.Y
	ystep := 1
	if start.Y >= end.Y {
		ystep = -1
	}
	for x := start.X; x < end.X; x++ {
		if steep {
			result = append(result, Pos{y, x})
		} else {
			result = append(result, Pos{x, y})
		}
		err += deltaY
		if 2*err >= deltaX {
			y += ystep
			err -= deltaX
		}
	}
	return result
}

// loadLevelFromFile - reads in and parses a level file
// properly associates each ascii character with a tile (not texture itself)
func loadLevelFromFile(filename string) *Level {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	levelLines := make([]string, 0)
	longestRow := 0
	index := 0
	for scanner.Scan() {
		levelLines = append(levelLines, scanner.Text())
		if len(levelLines[index]) > longestRow {
			longestRow = len(levelLines[index])
		}
		index++
	}

	level := &Level{}
	// TODO where we should init player?
	level.Player = &Player{}
	level.Player.Strength = 20
	level.Player.Hitpoints = 1000
	level.Player.Name = "Riley"
	level.Player.Symbol = '@'
	level.Player.Speed = 1.0
	level.Player.ActionPoints = 0
	level.Player.MaxBreath = 10
	level.Player.CurrentBreath = level.Player.MaxBreath
	level.Player.SightRange = 10

	level.Map = make([][]Tile, len(levelLines))
	level.Monsters = make(map[Pos]*Monster)
	level.Trees = make(map[Pos]Tile)
	level.Events = make([]string, 10) // 10 = number of events that fit on screen at a time
	level.Debug = make(map[Pos]bool)

	for i := range level.Map {
		level.Map[i] = make([]Tile, longestRow) // refactor to jagged array?
	}

	for y := 0; y < len(level.Map); y++ {
		line := levelLines[y]
		for x, c := range line {
			var t Tile
			switch c {
			case ' ', '\t', '\n', '\r':
				t.Symbol = Blank
			case '#':
				t.Symbol = StoneWall
			case '|':
				t.Symbol = ClosedDoor
			case '/':
				t.Symbol = OpenDoor
			case '.':
				t.Symbol = DirtFloor
			case ',':
				t.Symbol = Grass
			case '^':
				t.Symbol = Tree
				level.Trees[Pos{x, y}] = t
			case '~':
				t.Symbol = Water
			case '$':
				t.Symbol = Sand
			case '@':
				level.Player.Y = y
				level.Player.X = x
				t.Symbol = Pending
			case 'R':
				level.Monsters[Pos{x, y}] = NewRat(Pos{x, y})
				t.Symbol = Pending
			case 'S':
				level.Monsters[Pos{x, y}] = NewSpider(Pos{x, y})
				t.Symbol = Pending
			default:
				panic("Invalid Character in Map")
			}
			level.Map[y][x] = t
		}
	}

	// Handle pending tiles by setting floor to closest floor type found (or dirt for default)
	for y, row := range level.Map {
		for x, tile := range row {
			if tile.Symbol == Pending {
				level.Map[y][x] = level.bfsFloor(Pos{x, y})
			}
		}
	}

	level.lineOfSight()

	return level
}

// inRange checks that X pos and Y pos are within range of the map
func inRange(level *Level, pos Pos) bool {
	return pos.X < len(level.Map[0]) && pos.Y < len(level.Map) && pos.X >= 0 && pos.Y >= 0
}

// canWalk - determine if a tile should result in a collision or not
// possibly rename to be more general? (used in astar)
func canWalk(level *Level, pos Pos) bool {
	if inRange(level, pos) {
		t := level.Map[pos.Y][pos.X]
		switch t.Symbol {
		case StoneWall, ClosedDoor, Tree, Blank:
			return false
		}
		_, exists := level.Monsters[pos]
		return !exists
	}
	return false
}

func canSeeThrough(level *Level, pos Pos) bool {
	if inRange(level, pos) {
		t := level.Map[pos.Y][pos.X]
		switch t.Symbol {
		case StoneWall, ClosedDoor, Tree, Blank:
			return false
		default:
			return true
		}
	}
	return false
}

// checkDoor - open a closed door
func checkDoor(level *Level, pos Pos) {
	t := level.Map[pos.Y][pos.X]
	if t.Symbol == ClosedDoor {
		level.Map[pos.Y][pos.X].Symbol = OpenDoor
		level.lineOfSight()
	}
}

func (level *Level) resolveMovement(pos Pos) {
	monster, exists := level.Monsters[pos]
	if exists {
		level.Attack(&level.Player.Character, &monster.Character)
		if monster.Hitpoints <= 0 {
			delete(level.Monsters, monster.Pos)
			level.AddEvent(fmt.Sprintf("%s is dead", monster.Name))
		}
		if level.Player.Hitpoints <= 0 {
			level.AddEvent("Player died")
			sdl.Quit()
			os.Exit(1)
		}
	} else if canWalk(level, pos) {
		level.Player.Move(pos, level)
	} else {
		checkDoor(level, pos)
	}

	// Check if player is drowning
	if level.Map[level.Player.Pos.Y][level.Player.Pos.X].Symbol == '~' {
		level.AddEvent(fmt.Sprintf("Player has %d breath remaining", level.Player.CurrentBreath))
		level.Player.CurrentBreath -= 1
		if level.Player.CurrentBreath < 0 {
			level.AddEvent("Player died")
			sdl.Quit()
			os.Exit(1)
		}
	} else {
		level.Player.CurrentBreath = level.Player.MaxBreath
	}
}

func (player *Player) Move(to Pos, level *Level) {
	player.Pos = to
	for y, row := range level.Map {
		for x, _ := range row {
			level.Map[y][x].Visible = false
		}
	}
	level.lineOfSight()
}

// handleInput - takes an input and performs a game action
func (game *Game) handleInput(input *Input) {
	level := game.Level
	switch input.Typ {
	case Up:
		level.resolveMovement(Pos{level.Player.X, level.Player.Y - 1})
	case Down:
		level.resolveMovement(Pos{level.Player.X, level.Player.Y + 1})
	case Left:
		level.resolveMovement(Pos{level.Player.X - 1, level.Player.Y})
	case Right:
		level.resolveMovement(Pos{level.Player.X + 1, level.Player.Y})
	case CloseWindow:
		close(input.LevelChannel)
		chanIndex := 0
		for i, c := range game.LevelChans {
			if c == input.LevelChannel {
				chanIndex = i
				break
			}
		}
		// remove channel from slice
		game.LevelChans = append(game.LevelChans[:chanIndex], game.LevelChans[chanIndex+1:]...)
	}
}

// getNeighbors - returns an array containing the positions of each neighboring tile
func getNeighbors(level *Level, pos Pos) []Pos {
	neighbors := make([]Pos, 0, 4)
	left := Pos{pos.X - 1, pos.Y}
	right := Pos{pos.X + 1, pos.Y}
	up := Pos{pos.X, pos.Y - 1}
	down := Pos{pos.X, pos.Y + 1}

	if canWalk(level, right) {
		neighbors = append(neighbors, right)
	}
	if canWalk(level, left) {
		neighbors = append(neighbors, left)
	}
	if canWalk(level, up) {
		neighbors = append(neighbors, up)
	}
	if canWalk(level, down) {
		neighbors = append(neighbors, down)
	}

	return neighbors
}

func (pos *Pos) IsNextTo(tile Tile, level *Level) bool {
	left := Pos{pos.X - 1, pos.Y}
	right := Pos{pos.X + 1, pos.Y}
	up := Pos{pos.X, pos.Y - 1}
	down := Pos{pos.X, pos.Y + 1}
	if level.Map[left.X][left.Y] == tile ||
		level.Map[right.X][right.Y] == tile ||
		level.Map[up.X][up.Y] == tile ||
		level.Map[down.X][down.Y] == tile {
		return true
	}
	return false
}

func (pos *Pos) IsNextToPlayer(level *Level) bool {
	left := Pos{pos.X - 1, pos.Y}
	right := Pos{pos.X + 1, pos.Y}
	up := Pos{pos.X, pos.Y - 1}
	down := Pos{pos.X, pos.Y + 1}
	if left == level.Player.Pos ||
		right == level.Player.Pos ||
		up == level.Player.Pos ||
		down == level.Player.Pos {
		return true
	}
	return false
}

// bfs - classic breadth first search implementation
func (level *Level) bfsFloor(start Pos) Tile {
	frontier := make([]Pos, 0, 8)
	frontier = append(frontier, start)
	visited := make(map[Pos]bool)
	visited[start] = true

	for len(frontier) > 0 {
		current := frontier[0]
		currentTile := level.Map[current.Y][current.X]
		switch currentTile.Symbol {
		case DirtFloor:
			return Tile{DirtFloor, false}
		case Grass:
			return Tile{Grass, false}
		case Sand:
			return Tile{Sand, false}
		default:
		}
		// new slice starting from second element to the end
		frontier = frontier[1:]
		for _, next := range getNeighbors(level, current) {
			if !visited[next] {
				// add nodes not visited to queue
				frontier = append(frontier, next)
				visited[next] = true
			}
		}
	}
	return Tile{DirtFloor, false}
}

// astar - classic astar implementation
func (level *Level) astar(start Pos, goal Pos) []Pos {
	frontier := make(pqueue, 0, 8)
	frontier = frontier.push(start, 1)
	cameFrom := make(map[Pos]Pos)
	cameFrom[start] = start
	costSoFar := make(map[Pos]int)
	costSoFar[start] = 0
	var current Pos

	for len(frontier) > 0 {
		frontier, current = frontier.pop()

		// found path
		if current == goal {
			path := make([]Pos, 0)
			p := current
			for p != start {
				path = append(path, p)
				p = cameFrom[p]
			}
			path = append(path, p)
			// reverse path
			for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
				path[i], path[j] = path[j], path[i]
			}
			return path
		}

		for _, next := range getNeighbors(level, current) {
			newCost := costSoFar[current] + 1 // always 1 for now
			_, exists := costSoFar[next]
			if !exists || newCost < costSoFar[next] {
				costSoFar[next] = newCost
				xDist := int(math.Abs(float64(goal.X - next.X)))
				yDist := int(math.Abs(float64(goal.Y - next.Y)))
				priority := newCost + xDist + yDist
				frontier = frontier.push(next, priority)
				cameFrom[next] = current
			}
		}
	}
	return nil
}

// Run - contains main game loop
func (game *Game) Run() {
	for _, lchan := range game.LevelChans {
		lchan <- game.Level
	}

	for input := range game.InputChan {
		// quit game
		if input.Typ == QuitGame {
			return
		}

		game.handleInput(input)

		// move monsters towards player
		for _, monster := range game.Level.Monsters {
			monster.Update(game.Level)
		}

		// all windows have been closed
		if len(game.LevelChans) == 0 {
			return
		}

		// update game level
		for _, lchan := range game.LevelChans {
			lchan <- game.Level
		}
	}
}
