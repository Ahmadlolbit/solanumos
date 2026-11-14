package theme

import (
	"fmt"
	"math"

	"github.com/Ahmadlolbit/solanumos/internal/follow"
)

const (
	black       = "#000000"
	white       = "#ffffff"
	hueChroma   = 0.11
	searchSteps = 40
	mixStep     = 0.005
)

type rgb [3]float64

type hueAngle struct {
	key string
	at  float64
}

var (
	degree     = divided(math.Pi, 180)
	gammaPower = divided(1, 2.4)
	gamutTop   = 1 + gamutSlack()
	hueAngles  = []hueAngle{{"red", 25}, {"orange", 55}, {"yellow", 95}, {"green", 145}, {"cyan", 200}, {"blue", 250}, {"magenta", 330}}
)

func divided(top, bottom float64) float64 {
	return top / bottom
}

func gamutSlack() float64 {
	return 1e-4
}

func floats(color string) rgb {
	channels := follow.Channels(color)
	return rgb{float64(channels[0]) / 255, float64(channels[1]) / 255, float64(channels[2]) / 255}
}

func (color rgb) hex() string {
	text := "#"
	for _, value := range color {
		text += fmt.Sprintf("%02x", int(min(255, max(0, math.RoundToEven(value*255)))))
	}
	return text
}

func linear(value float64) float64 {
	if value <= 0.04045 {
		return value / 12.92
	}
	return math.Pow((value+0.055)/1.055, 2.4)
}

func gamma(value float64) float64 {
	if value <= 0.0031308 {
		return 12.92 * value
	}
	return float64(1.055*math.Pow(value, gammaPower)) - 0.055
}

func luminance(color string) float64 {
	value := floats(color)
	return float64(0.2126*linear(value[0])) + float64(0.7152*linear(value[1])) + float64(0.0722*linear(value[2]))
}

func Contrast(one, other string) float64 {
	first, second := luminance(one), luminance(other)
	if first < second {
		first, second = second, first
	}
	return (first + 0.05) / (second + 0.05)
}

func Dark(background string) bool {
	return Contrast(background, white) >= Contrast(background, black)
}

func blend(from, to string, share float64) string {
	one, other := floats(from), floats(to)
	var mixed rgb
	for index := range mixed {
		mixed[index] = one[index] + float64((other[index]-one[index])*share)
	}
	return mixed.hex()
}

func oklab(lightness, chroma, hue float64) rgb {
	radians := hue * degree
	a, b := chroma*math.Cos(radians), chroma*math.Sin(radians)
	l := math.Pow(lightness+float64(0.3963377774*a)+float64(0.2158037573*b), 3)
	m := math.Pow(lightness-float64(0.1055613458*a)-float64(0.0638541728*b), 3)
	s := math.Pow(lightness-float64(0.0894841775*a)-float64(1.2914855480*b), 3)
	return rgb{
		float64(4.0767416621*l) - float64(3.3077115913*m) + float64(0.2309699292*s),
		float64(-1.2684380046*l) + float64(2.6097574011*m) - float64(0.3413193965*s),
		float64(-0.0041960863*l) - float64(0.7034186147*m) + float64(1.7076147010*s),
	}
}

func inGamut(color rgb) bool {
	for _, value := range color {
		if value < -gamutSlack() || value > gamutTop {
			return false
		}
	}
	return true
}

func oklch(lightness, chroma, hue float64) string {
	for chroma > 0 && !inGamut(oklab(lightness, chroma, hue)) {
		chroma -= 0.002
	}
	color := oklab(lightness, max(chroma, 0), hue)
	for index, value := range color {
		color[index] = gamma(min(1, max(0, value)))
	}
	return color.hex()
}

func hueFor(background string, at, chroma, target float64) string {
	low, high := 0.0, 1.0
	dark := Dark(background)
	for range searchSteps {
		middle := (low + high) / 2
		if (Contrast(oklch(middle, chroma, at), background) < target) == dark {
			low = middle
		} else {
			high = middle
		}
	}
	if dark {
		return oklch(high, chroma, at)
	}
	return oklch(low, chroma, at)
}

func firstMix(background, foreground string, target float64) string {
	share := 0.0
	for share < 1 && Contrast(blend(background, foreground, share), background) < target {
		share += mixStep
	}
	return blend(background, foreground, share)
}

func hueTargets(key string) (float64, float64) {
	if key == "yellow" {
		return 7.5, 9.5
	}
	return 6, 8
}

func modeOf(background string) string {
	if Dark(background) {
		return "dark"
	}
	return "light"
}
