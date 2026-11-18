package qr

import (
	"errors"
	"strings"

	"github.com/Ahmadlolbit/solanumos/internal/run"
)

const quiet = 2

func Matrix(text string) ([][]bool, error) {
	output, _, err := run.CaptureFed(run.Quick, text, "qrencode", "-t", "ASCII", "-m", "0", "-l", "L", "-o", "-")
	if err != nil {
		return nil, err
	}
	return Parse(output)
}

func Parse(output string) ([][]bool, error) {
	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
	matrix := make([][]bool, len(lines))
	for y, line := range lines {
		if len(line) != 2*len(lines) {
			return nil, errors.New("qrencode drew a code that isn't square")
		}
		matrix[y] = make([]bool, len(lines))
		for x := range matrix[y] {
			matrix[y][x] = line[2*x] == '#'
		}
	}
	return matrix, nil
}

func light(matrix [][]bool, x, y int) bool {
	x, y = x-quiet, y-quiet
	if y < 0 || x < 0 || y >= len(matrix) || x >= len(matrix) {
		return true
	}
	return !matrix[y][x]
}

func Halves(matrix [][]bool) []string {
	size := len(matrix) + 2*quiet
	glyphs := map[[2]bool]string{{true, true}: "█", {true, false}: "▀", {false, true}: "▄", {false, false}: " "}
	var lines []string
	for y := 0; y < size; y += 2 {
		var line strings.Builder
		for x := range size {
			line.WriteString(glyphs[[2]bool{light(matrix, x, y), y+1 < size && light(matrix, x, y+1)}])
		}
		lines = append(lines, line.String())
	}
	return lines
}

func Blocks(matrix [][]bool) []string {
	size := len(matrix) + 2*quiet
	lines := make([]string, 0, size)
	for y := range size {
		var line strings.Builder
		for x := range size {
			if light(matrix, x, y) {
				line.WriteString("██")
			} else {
				line.WriteString("  ")
			}
		}
		lines = append(lines, line.String())
	}
	return lines
}

func Lines(text string, console bool) ([]string, error) {
	matrix, err := Matrix(text)
	if err != nil {
		return nil, err
	}
	if console {
		return Blocks(matrix), nil
	}
	return Halves(matrix), nil
}
