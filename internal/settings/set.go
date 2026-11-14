package settings

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/look"
	"github.com/Ahmadlolbit/solanumos/internal/run"
	"github.com/Ahmadlolbit/solanumos/internal/theme"
)

const Helper = "solanum-settings-helper"

var (
	clockTime = regexp.MustCompile(`^([01][0-9]|2[0-3]):[0-5][0-9]$`)
	Runner    = func(name string, args ...string) int { return run.Status(name, args...) }
)

func Find(rows []Row, name string) (Row, bool) {
	for _, row := range rows {
		if row.Name == name {
			return row, true
		}
	}
	return Row{}, false
}

func listed(row Row, value string) bool {
	values := make([]string, 0, len(row.Choices))
	for _, choice := range row.Choices {
		values = append(values, choice.Value)
	}
	if !row.Multiple {
		return slices.Contains(values, value) || (row.Name == "accent" && theme.IsHex(value)) || (strings.HasSuffix(row.Name, "font") && theme.ValidFont(value))
	}
	return !slices.ContainsFunc(strings.Fields(value), func(word string) bool { return !slices.Contains(values, word) })
}

func within(row Row, value string) bool {
	number, err := strconv.ParseFloat(value, 64)
	return err == nil && number >= row.Min && number <= row.Max
}

var (
	checks = map[string]func(Row, string) bool{
		Toggle: func(_ Row, value string) bool { return value == "yes" || value == "no" },
		Pick:   func(row Row, value string) bool { return row.Whose == Computer || listed(row, value) },
		Slide:  func(row Row, value string) bool { return row.Whose == Computer || within(row, value) },
		Time:   func(_ Row, value string) bool { return clockTime.MatchString(value) },
		Text:   func(Row, string) bool { return true },
		Button: func(Row, string) bool { return false },
	}
	problems = map[string]string{
		Toggle: "%s takes yes or no", Pick: "%s takes one of its choices; see solanum settings list", Slide: "%s takes a number in its range",
		Time: "%s takes a time like 07:00", Button: "%s opens something; it has no value to set",
	}
)

func Valid(row Row, value string) error {
	if check, known := checks[row.Kind]; !known || !check(row, value) {
		return fmt.Errorf(problems[row.Kind], row.Name)
	}
	if strings.ContainsAny(value, "\n\r") {
		return fmt.Errorf("%s takes one line", row.Name)
	}
	if option, found := Option(row.Name); found && row.Whose == Computer && !option.Valid(value) {
		return errors.New(option.Refusal(value))
	}
	return nil
}

func barJSON(key, value string) (json.RawMessage, error) {
	switch key {
	case "workspaces":
		numbers := []int{}
		for _, word := range strings.Fields(value) {
			number, err := strconv.Atoi(word)
			if err != nil || number < 1 {
				return nil, fmt.Errorf("bar.workspaces takes workspace numbers like 1 2 3")
			}
			numbers = append(numbers, number)
		}
		return json.Marshal(numbers)
	case "left", "center", "right":
		return json.Marshal(strings.Fields(value))
	}
	return json.Marshal(value)
}

func setBar(key, value string) (int, error) {
	if key == "preset" || key == "position" {
		return Runner(run.Sibling("solanum-look"), map[string]string{"preset": "bar", "position": "position"}[key], value), nil
	}
	raw, err := barJSON(key, value)
	if err == nil {
		err = look.SetBar(shellFile(), key, raw)
	}
	if err != nil {
		return 1, err
	}
	return 0, nil
}

func Set(name, value string) (int, error) {
	row, found := Find(List().Rows, name)
	if !found {
		return 2, fmt.Errorf("there's no setting named %s; see solanum settings list", name)
	}
	if err := Valid(row, value); err != nil {
		return 2, err
	}
	switch {
	case row.Whose == Computer:
		return Runner("pkexec", filepath.Join(env.BinDir(), Helper), name, value), nil
	case strings.HasPrefix(name, "bar."):
		return setBar(strings.TrimPrefix(name, "bar."), value)
	case name == "logo":
		return Runner(run.Sibling("solanum-background"), map[string]string{"yes": "logo", "no": "next"}[value]), nil
	}
	return Runner(run.Sibling("solanum-look"), lookArgs(name, value, look.Open().Get)...), nil
}
