// Riley Mulford April 2019
package game

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"sort"
	"time"
)

type GameUI interface {
	Draw(*Level)
	GetInput() *Input
}

const (
	None InputType = iota
	Up
	Down
	Left
	Right
	Quit
	Search // TODO remove
)

type InputType int

type Input struct {
	Typ InputType
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

type priorityPos struct {
	Pos
	priority int
}

type priorityArray []priorityPos

// satisfy go sort interface
func (p priorityArray) Len() int           { return len(p) }
func (p priorityArray) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p priorityArray) Less(i, j int) bool { return p[i].priority < p[j].priority }

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
func handleInput(ui GameUI, level *Level, input *Input) {
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
		astar(ui, level, p.Pos, Pos{p.X + 2, p.Y + 1})
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
func bfs(ui GameUI, level *Level, start Pos) {
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
				ui.Draw(level)
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
}

// astar - classic astar implementation
func astar(ui GameUI, level *Level, start Pos, goal Pos) {
	frontier := make(priorityArray, 0, 8)
	frontier = append(frontier, priorityPos{start, 1})
	cameFrom := make(map[Pos]Pos)
	cameFrom[start] = start
	costSoFar := make(map[Pos]int)
	costSoFar[start] = 0
	level.Debug = make(map[Pos]bool)

	for len(frontier) > 0 {
		sort.Stable(frontier) // TODO fix slow priority queue
		current := frontier[0]

		// found path
		if current.Pos == goal {
			p := current.Pos
			for p != start {
				level.Debug[p] = true
				ui.Draw(level)
				time.Sleep(100 * time.Millisecond)
				p = cameFrom[p]
			}
			level.Debug[p] = true
			ui.Draw(level)
			time.Sleep(100 * time.Millisecond)
			fmt.Println("done!")
			break
		}

		frontier = frontier[1:]
		for _, next := range getNeighbors(level, current.Pos) {
			newCost := costSoFar[current.Pos] + 1 // always 1 for now
			_, exists := costSoFar[next]
			if !exists || newCost < costSoFar[next] {
				costSoFar[next] = newCost
				xDist := int(math.Abs(float64(goal.X - next.X)))
				yDist := int(math.Abs(float64(goal.Y - next.Y)))
				priority := newCost + xDist + yDist
				frontier = append(frontier, priorityPos{next, priority})
				cameFrom[next] = current.Pos
			}
		}
	}
}

// Run - contains main game loop
func Run(ui GameUI) {
	level := loadLevelFromFile("game/maps/level1.map")
	for {
		ui.Draw(level)
		input := ui.GetInput()

		if input != nil && input.Typ == Quit {
			return
		}

		handleInput(ui, level, input)
	}
}
