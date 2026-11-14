package wallpaper

import "github.com/Ahmadlolbit/solanumos/internal/hypr"

const largest = 8192

func Size() (int, int) {
	width, height, ok := hypr.LargestScreen()
	if !ok || width > largest || height > largest {
		return Width, Height
	}
	return width, height
}
