package game

import (
	"fmt"
	"math"
	"os"

	"github.com/veandco/go-sdl2/sdl"
)

type Monster struct {
	Character
}

func NewRat(p Pos) *Monster {
	// TODO figure out why this doesnt work
	//return &Monster{Pos: p, Symbol: 'R', Name: "Rat", Hitpoints: 5, Strength: 5, Speed: 2.0, ActionPoints: 0.0}
	monster := &Monster{}
	monster.Pos = p
	monster.Symbol = 'R'
	monster.Name = "Rat"
	monster.Hitpoints = 50
	monster.Strength = 5
	monster.Speed = 2.0
	monster.ActionPoints = 0.0
	return monster
}

func NewSpider(p Pos) *Monster {
	//return &Monster{p, 'S', "Spider", 10, 10, 1.0, 0.0}
	monster := &Monster{}
	monster.Pos = p
	monster.Symbol = 'S'
	monster.Name = "Spider"
	monster.Hitpoints = 100
	monster.Strength = 10
	monster.Speed = 1.0
	monster.ActionPoints = 0.0
	return monster
}

func (m *Monster) Update(level *Level) {
	m.ActionPoints += m.Speed
	path := level.astar(m.Pos, level.Player.Pos)
	cost := math.Trunc((math.Min(m.ActionPoints, float64(len(path)-1))))
	m.Move(path[int(cost)], level)
	m.ActionPoints -= cost
}

func (m *Monster) Move(to Pos, level *Level) {
	_, exists := level.Monsters[to]
	// TODO check if tile being moved to is valid (walls)
	if !exists && to != level.Player.Pos {
		delete(level.Monsters, m.Pos)
		level.Monsters[to] = m
		m.Pos = to
	} else {
		if m.Pos.IsNextToPlayer(level) {
			Attack(&m.Character, &level.Player.Character)
			// monster died
			if m.Hitpoints <= 0 {
				delete(level.Monsters, m.Pos)
			}
			// player died
			if level.Player.Hitpoints <= 0 {
				fmt.Println("Player died")
				sdl.Quit()
				os.Exit(1)
			}
		}
	}
}
