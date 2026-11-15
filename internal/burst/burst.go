package burst

import (
	"bytes"
	_ "embed"
	"image"
	"image/color"
	"image/png"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	Size = 32
	tick = time.Second / 60
	mark = "logo6"
)

type Frame [Size]uint32

type shown struct {
	ticks  int
	name   string
	pixels Frame
}

type Step struct {
	Frame Frame
	Hold  time.Duration
}

//go:embed flipbook.txt
var flipbook string

var loop = sync.OnceValue(func() []shown { return read(flipbook) })

func read(text string) []shown {
	var frames []shown
	for _, line := range strings.Split(strings.TrimSpace(text), "\n") {
		fields := strings.Fields(line)
		if len(fields) != Size+2 {
			continue
		}
		ticks, err := strconv.Atoi(fields[0])
		if err != nil || ticks < 1 {
			continue
		}
		frame := shown{ticks: ticks, name: fields[1]}
		for row, word := range fields[2:] {
			value, _ := strconv.ParseUint(word, 16, 32)
			frame.pixels[row] = uint32(value)
		}
		frames = append(frames, frame)
	}
	return frames
}

func Flipbook() string {
	return flipbook
}

func Length() time.Duration {
	total := 0
	for _, frame := range loop() {
		total += frame.ticks
	}
	return time.Duration(total) * tick
}

func markStart() time.Duration {
	start := 0
	for _, frame := range loop() {
		if frame.name == mark {
			break
		}
		start += frame.ticks
	}
	return time.Duration(start) * tick
}

func Steps() []Step {
	steps := make([]Step, 0, len(loop()))
	for _, frame := range loop() {
		steps = append(steps, Step{Frame: frame.pixels, Hold: time.Duration(frame.ticks) * tick})
	}
	return steps
}

func Mark() int {
	return max(slices.IndexFunc(loop(), func(frame shown) bool { return frame.name == mark }), 0)
}

func Still() Frame {
	for _, frame := range loop() {
		if frame.name == mark {
			return frame.pixels
		}
	}
	return Frame{}
}

func At(elapsed time.Duration) (Frame, time.Duration) {
	length := Length()
	if length == 0 {
		return Frame{}, time.Hour
	}
	position := (max(elapsed, 0) + markStart()) % length
	at := int(position / tick)
	for _, frame := range loop() {
		if at < frame.ticks {
			return frame.pixels, time.Duration(frame.ticks-at)*tick - position%tick
		}
		at -= frame.ticks
	}
	return Still(), length
}

func cell(top, bottom bool) rune {
	switch {
	case top && bottom:
		return '█'
	case top:
		return '▀'
	case bottom:
		return '▄'
	}
	return ' '
}

func Lines(frame Frame) []string {
	lines := make([]string, 0, Size/2)
	for row := 0; row < Size; row += 2 {
		var line strings.Builder
		for column := Size - 1; column >= 0; column-- {
			line.WriteRune(cell(frame[row]>>column&1 == 1, frame[row+1]>>column&1 == 1))
		}
		lines = append(lines, line.String())
	}
	return lines
}

func PNG(scale int, paint color.Color) ([]byte, error) {
	still := Still()
	side := Size * scale
	picture := image.NewNRGBA(image.Rect(0, 0, side, side))
	for at := range side * side {
		x, y := at%side, at/side
		if still[y/scale]>>(Size-1-x/scale)&1 == 1 {
			picture.Set(x, y, paint)
		}
	}
	var out bytes.Buffer
	err := png.Encode(&out, picture)
	return out.Bytes(), err
}

func Picture(frame Frame, scale int, paint color.Color) ([]byte, error) {
	side := Size * scale
	picture := image.NewPaletted(image.Rect(0, 0, side, side), color.Palette{color.NRGBA{}, paint})
	for at := range side * side {
		x, y := at%side, at/side
		if frame[y/scale]>>(Size-1-x/scale)&1 == 1 {
			picture.SetColorIndex(x, y, 1)
		}
	}
	var out bytes.Buffer
	err := png.Encode(&out, picture)
	return out.Bytes(), err
}
