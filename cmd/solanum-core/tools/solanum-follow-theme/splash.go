package followtheme

import (
	"fmt"
	"image/color"
	"path/filepath"
	"strings"
	"time"

	"github.com/Ahmadlolbit/solanumos/internal/burst"
)

const refreshesPerSecond = 50

func holds(steps []burst.Step) string {
	parts := make([]string, 0, len(steps))
	for index, step := range steps {
		refreshes := max(int((step.Hold*refreshesPerSecond+time.Second/2)/time.Second), 1)
		parts = append(parts, fmt.Sprintf("hold[%d] = %d;", index, refreshes))
	}
	return strings.Join(parts, " ")
}

func logoFiles(scale int, steps []burst.Step, paint color.Color) ([][2]string, error) {
	still, err := burst.PNG(scale, paint)
	if err != nil {
		return nil, err
	}
	files := [][2]string{{filepath.Join(plymouthDir, fmt.Sprintf("burst-%d.png", scale)), string(still)}}
	for index, step := range steps {
		picture, err := burst.Picture(step.Frame, scale, paint)
		if err != nil {
			return nil, err
		}
		files = append(files, [2]string{filepath.Join(plymouthDir, fmt.Sprintf("burst-%d-%d.png", scale, index)), string(picture)})
	}
	return files, nil
}
