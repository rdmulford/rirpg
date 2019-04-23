// Riley Mulford April 2019
package game

import (
	"bufio"
	"math"
	"os"
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

type Tile rune

const (
	StoneWall  Tile = '#'
	DirtFloor  Tile = '.'
	ClosedDoor Tile = '|'
	OpenDoor   Tile = '/'
	Blank      Tile = 0
	Pending    Tile = -1
)

type Pos struct {
	X, Y int
}

type Entity struct {
	Pos
}

type Player struct {
	Entity
}

type Level struct {
	Map    [][]Tile
	Player Player
	Debug  map[Pos]bool
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
		index += 1
	}
	level := &Level{}
	level.Map = make([][]Tile, len(levelLines))
	for i := range level.Map {
		level.Map[i] = make([]Tile, longestRow) // refactor to jagged array?
	}

	for y := 0; y < len(level.Map); y += 1 {
		line := levelLines[y]
		for x, c := range line {
			var t Tile
			switch c {
			case ' ', '\t', '\n', '\r':
				t = Blank
			case '#':
				t = StoneWall
			case '|':
				t = ClosedDoor
			case '/':
				t = OpenDoor
			case '.':
				t = DirtFloor
			case 'P':
				level.Player.Y = y
				level.Player.X = x
				t = Pending
			default:
				panic("Invalid Character in Map")
			}
			level.Map[y][x] = t
		}
	}

	for y, row := range level.Map {
		for x, tile := range row {
			if tile == Pending {
			SearchLoop:
				for searchX := x - 1; searchX < x+1; searchX += 1 {
					for searchY := y - 1; searchY < y+1; searchY += 1 {
						searchTile := level.Map[searchY][searchX]
						switch searchTile {
						case DirtFloor:
							level.Map[y][x] = DirtFloor
							break SearchLoop
						}
					}
				}
			}
		}
	}

	return level
}

// canWalk - determine if a tile should result in a collision or not
// possibly rename to be more general? (used in astar)
func canWalk(level *Level, pos Pos) bool {
	t := level.Map[pos.Y][pos.X]
	switch t {
	case StoneWall, ClosedDoor, Blank:
		return false
	default:
		return true
	}
}

// checkDoor - open a closed door
func checkDoor(level *Level, pos Pos) {
	t := level.Map[pos.Y][pos.X]
	if t == ClosedDoor {
		level.Map[pos.Y][pos.X] = OpenDoor
	}
}

// handleInput - takes an input and performs a game action
func (game *Game) handleInput(input *Input) {
	level := game.Level
	p := level.Player
	switch input.Typ {
	case Up:
		if canWalk(level, Pos{p.X, p.Y - 1}) {
			level.Player.Y--
		} else {
			checkDoor(level, Pos{p.X, p.Y - 1})
		}
	case Down:
		if canWalk(level, Pos{p.X, p.Y + 1}) {
			level.Player.Y++
		} else {
			checkDoor(level, Pos{p.X, p.Y + 1})
		}
	case Left:
		if canWalk(level, Pos{p.X - 1, p.Y}) {
			level.Player.X--
		} else {
			checkDoor(level, Pos{p.X - 1, p.Y})
		}
	case Right:
		if canWalk(level, Pos{p.X + 1, p.Y}) {
			level.Player.X++
		} else {
			checkDoor(level, Pos{p.X + 1, p.Y})
		}
	case Search:
		game.astar(level.Player.Pos, Pos{3, 2})
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

// bfs - classic breadth first search implementation
func (game *Game) bfs(start Pos) {
	level := game.Level
	frontier := make([]Pos, 0, 8)
	frontier = append(frontier, start)
	visited := make(map[Pos]bool)
	visited[start] = true
	level.Debug = visited

	for len(frontier) > 0 {
		current := frontier[0]
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
}

// astar - classic astar implementation
func (game *Game) astar(start Pos, goal Pos) []Pos {
	level := game.Level
	frontier := make(pqueue, 0, 8)
	frontier = frontier.push(start, 1)
	cameFrom := make(map[Pos]Pos)
	cameFrom[start] = start
	costSoFar := make(map[Pos]int)
	costSoFar[start] = 0
	level.Debug = make(map[Pos]bool)
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
			// DEBUG
			for _, pos := range path {
				level.Debug[pos] = true
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
