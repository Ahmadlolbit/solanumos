package weather

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Ahmadlolbit/solanumos/internal/atomic"
	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/look"
	"github.com/Ahmadlolbit/solanumos/internal/weather"
)

const usage = `Usage: solanum-weather [update]

Prints the weather saved in ~/.cache/solanum/weather.json; update
fetches it again from Open-Meteo (no account) for the look settings'
location: your time zone's main city, or solanum-look location.
solanum-look weather on fetches it every 30 minutes for the bar.
`

func cache() string {
	return filepath.Join(env.CacheHome(), "solanum", "weather.json")
}

func update() int {
	settings := look.Open()
	coordinates, ok := settings.Coordinates()
	if !ok {
		fmt.Fprintln(os.Stderr, "no location: set one with solanum-look location LATITUDE,LONGITUDE")
		return 1
	}
	url := weather.URL(coordinates, settings.Get("weather_units"))
	if err := os.MkdirAll(filepath.Dir(cache()), 0o777); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	data, err := weather.Fetch(url)
	if err != nil {
		fmt.Fprintln(os.Stderr, "could not fetch the weather")
		return 1
	}
	if data, err = weather.Stamp(data, time.Now().Unix()); err != nil {
		fmt.Fprintf(os.Stderr, "could not read the weather: %v\n", err)
		return 1
	}
	if err := atomic.Write(cache(), data, 0o666); err != nil {
		fmt.Fprintf(os.Stderr, "could not save the weather: %v\n", err)
		return 1
	}
	return 0
}

func show() int {
	if info, err := os.Stat(cache()); err != nil || info.Size() == 0 {
		fmt.Fprintln(os.Stderr, "no weather yet: solanum-weather update")
		return 1
	}
	data, err := os.ReadFile(cache())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	summary, err := weather.Summary(data)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 5
	}
	fmt.Println(summary)
	return 0
}

func Main() {
	switch env.Arg(1) {
	case "":
		os.Exit(show())
	case "update":
		os.Exit(update())
	case "-h", "--help":
		fmt.Print(usage)
	default:
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
}
