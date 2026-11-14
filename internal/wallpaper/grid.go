package wallpaper

type cells struct {
	left, top, columns, rows int
	values                   []float64
	value                    func(i, j int) float64
}

func remembered(left, top, right, bottom int, value func(i, j int) float64) cells {
	grid := cells{left, top, right - left + 1, bottom - top + 1, nil, value}
	grid.values = make([]float64, grid.columns*grid.rows)
	for at := range grid.values {
		grid.values[at] = value(left+at%grid.columns, top+at/grid.columns)
	}
	return grid
}

func (grid cells) at(i, j int) float64 {
	x, y := i-grid.left, j-grid.top
	if x < 0 || y < 0 || x >= grid.columns || y >= grid.rows {
		return grid.value(i, j)
	}
	return grid.values[y*grid.columns+x]
}
