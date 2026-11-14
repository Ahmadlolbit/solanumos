package wallpaper

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Ahmadlolbit/solanumos/internal/fsutil"
	"github.com/Ahmadlolbit/solanumos/internal/theme"
)

const (
	lifeRows   = 90
	lifeSown   = 35
	lifeWarmup = 40
	lifeInk    = 50
	lifeStream = 0x9e3779b97f4a7c15
)

type lifePlace struct{ pitch, cols, rows, left, top int }

type Soup struct {
	Cols, Rows int
	cells      []bool
}

func lifeAt(width, height int) lifePlace {
	pitch := max(1, int(math.Round(float64(height)/lifeRows)))
	cols, rows := max(1, width/pitch), max(1, height/pitch)
	return lifePlace{pitch, cols, rows, (width - cols*pitch) / 2, (height - rows*pitch) / 2}
}

func Sow(seed uint32, cols, rows int) Soup {
	random := rand.New(rand.NewPCG(uint64(seed), lifeStream))
	soup := Soup{cols, rows, make([]bool, cols*rows)}
	for at := range soup.cells {
		soup.cells[at] = random.IntN(100) < lifeSown
	}
	for range lifeWarmup {
		soup = soup.next()
	}
	return soup
}

func (soup Soup) Alive(x, y int) bool {
	return soup.cells[(y%soup.Rows+soup.Rows)%soup.Rows*soup.Cols+(x%soup.Cols+soup.Cols)%soup.Cols]
}

func (soup Soup) neighbours(x, y int) int {
	count := 0
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if (dx != 0 || dy != 0) && soup.Alive(x+dx, y+dy) {
				count++
			}
		}
	}
	return count
}

func (soup Soup) next() Soup {
	after := Soup{soup.Cols, soup.Rows, make([]bool, len(soup.cells))}
	for at := range after.cells {
		count := soup.neighbours(at%soup.Cols, at/soup.Cols)
		after.cells[at] = count == 3 || count == 2 && soup.cells[at]
	}
	return after
}

func (soup Soup) String() string {
	var text strings.Builder
	for at, alive := range soup.cells {
		digit := byte('0')
		if alive {
			digit = '1'
		}
		text.WriteByte(digit)
		if at%soup.Cols == soup.Cols-1 {
			text.WriteByte('\n')
		}
	}
	return text.String()
}

func blend(from, to rgb, percent int) color.RGBA {
	part := func(one, other float64) uint8 {
		return uint8((int(one)*(100-percent) + int(other)*percent + 50) / 100)
	}
	return color.RGBA{part(from.r, to.r), part(from.g, to.g), part(from.b, to.b), 255}
}

func lifeImage(colors theme.Colors, soup Soup, place lifePlace, width, height int) *image.Paletted {
	base, ink := pick(colors, "background"), pick(colors, "foreground")
	img := image.NewPaletted(image.Rect(0, 0, width, height), color.Palette{blend(base, ink, 0), blend(base, ink, lifeInk)})
	for at, alive := range soup.cells {
		if !alive {
			continue
		}
		left, top := place.left+at%soup.Cols*place.pitch, place.top+at/soup.Cols*place.pitch
		for y := max(0, top); y < min(height, top+place.pitch); y++ {
			row := img.Pix[y*img.Stride:]
			for x := max(0, left); x < min(width, left+place.pitch); x++ {
				row[x] = 1
			}
		}
	}
	return img
}

func NewSeed() uint32 {
	return rand.Uint32()
}

func LifeStill(colors theme.Colors, seed uint32) (string, error) {
	width, height := Size()
	name := fmt.Sprintf("%08x-%08x-%dx%d.png", seed, Seed(colors["background"]+colors["foreground"]), width, height)
	path := filepath.Join(folder("life"), name)
	if fsutil.IsFile(path) {
		return path, nil
	}
	place := lifeAt(width, height)
	var out bytes.Buffer
	if err := png.Encode(&out, lifeImage(colors, Sow(seed, place.cols, place.rows), place, width, height)); err != nil {
		return "", err
	}
	return keep(path, out.Bytes())
}

func IsLife(path string) bool {
	return inFolder(path, "life")
}

func LifeSeed(link string) (uint32, bool) {
	target, err := os.Readlink(link)
	if err != nil || !IsLife(target) {
		return 0, false
	}
	field, _, _ := strings.Cut(filepath.Base(target), "-")
	seed, err := strconv.ParseUint(field, 16, 32)
	return uint32(seed), err == nil
}
