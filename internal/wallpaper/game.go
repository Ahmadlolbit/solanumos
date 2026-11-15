package wallpaper

import (
	"fmt"
	"math/rand/v2"
	"slices"

	"github.com/Ahmadlolbit/solanumos/internal/theme"
)

const (
	lifeTrail   = 3
	lifeCycle   = 16
	lifeRepeats = 60
)

var lifeInks = [lifeTrail + 2]int{lifeInk, 30, 16, 7, 0}

type Game struct {
	Generation, Soups, Population int
	soup                          Soup
	ages                          []uint8
	hashes                        [lifeCycle]uint64
	counts                        [2 * lifeRepeats]int
	same                          int
	settled                       bool
}

func NewGame(seed uint32, cols, rows int) *Game {
	soup := Sow(seed, cols, rows)
	game := &Game{Soups: 1, soup: soup, ages: make([]uint8, len(soup.cells))}
	for at, alive := range soup.cells {
		game.ages[at] = lifeTrail + 1
		if alive {
			game.ages[at] = 0
			game.Population++
		}
	}
	game.forget()
	return game
}

func (game *Game) forget() {
	game.hashes = [lifeCycle]uint64{}
	for at := range game.counts {
		game.counts[at] = -1
	}
	game.same = 0
}

func (game *Game) Step() {
	if game.settled {
		game.Sow(rand.Uint32())
		return
	}
	game.settle(game.soup.next())
}

func (game *Game) Sow(seed uint32) {
	game.Soups++
	game.forget()
	game.settle(Sow(seed, game.soup.Cols, game.soup.Rows))
}

func (game *Game) settle(next Soup) {
	population, hash := 0, uint64(14695981039346656037)
	for at, alive := range next.cells {
		switch {
		case alive:
			game.ages[at] = 0
			population++
			hash = (hash ^ uint64(at)) * 1099511628211
		case game.ages[at] <= lifeTrail:
			game.ages[at]++
		}
	}
	game.soup, game.Population = next, population
	game.Generation++
	seen := slices.Contains(game.hashes[:], hash)
	game.hashes[game.Generation%lifeCycle] = hash
	at := game.Generation % len(game.counts)
	game.same++
	if game.counts[(at+lifeRepeats)%len(game.counts)] != population {
		game.same = 0
	}
	game.counts[at] = population
	game.settled = population == 0 || seen || game.same >= lifeRepeats
}

func (game *Game) Frame(colors theme.Colors) []byte {
	base, ink := pick(colors, "background"), pick(colors, "foreground")
	var paints [lifeTrail + 2]color3
	for k, percent := range lifeInks {
		mixed := blend(base, ink, percent)
		paints[k] = color3{mixed.R, mixed.G, mixed.B}
	}
	header := fmt.Sprintf("P6\n%d %d\n255\n", game.soup.Cols, game.soup.Rows)
	frame := make([]byte, len(header), len(header)+3*len(game.ages))
	copy(frame, header)
	for _, age := range game.ages {
		paint := paints[min(int(age), lifeTrail+1)]
		frame = append(frame, paint[0], paint[1], paint[2])
	}
	return frame
}

type color3 [3]uint8
