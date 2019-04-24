package game

type Monster struct {
	Symbol    rune
	Name      string
	Hitpoints int
	Strength  int
	Speed     float64
}

func NewRat() *Monster {
	return &Monster{'R', "Rat", 5, 5, 2.0}
}

func NewSpider() *Monster {
	return &Monster{'S', "Spider", 10, 10, 1.0}
}
