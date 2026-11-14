package look

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/Ahmadlolbit/solanumos/internal/atomic"

	"github.com/Ahmadlolbit/solanumos/internal/ordered"
	"github.com/Ahmadlolbit/solanumos/internal/run"
)

const shellPath = "/usr/share/solanum/shell"

func CallShell(target, function string, args ...string) {
	run.Quiet(run.Quick, "qs", append([]string{"-p", shellPath, "ipc", "call", target, function}, args...)...)
}

func ReloadShell() {
	CallShell("theme", "reload")
}

type barSettings struct {
	Bar struct {
		Preset   string `json:"preset"`
		Position string `json:"position"`
	} `json:"bar"`
}

func Preset(data []byte) string {
	var shell barSettings
	if json.Unmarshal(unmarked(data), &shell) != nil {
		return ""
	}
	return shell.Bar.Preset
}

func Position(data []byte) string {
	var shell barSettings
	if json.Unmarshal(unmarked(data), &shell) != nil {
		return ""
	}
	return shell.Bar.Position
}

func WithPreset(data []byte, preset string) ([]byte, error) {
	return withBar(data, "preset", preset)
}

func WithPosition(data []byte, position string) ([]byte, error) {
	return withBar(data, "position", position)
}

func withBar(data []byte, key, value string) ([]byte, error) {
	name, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return WithBar(data, key, name)
}

func WithBar(data []byte, key string, value json.RawMessage) ([]byte, error) {
	document, err := parseOrEmpty(unmarked(data))
	if err != nil {
		return nil, err
	}
	raw, _ := document.Get("bar")
	bar, err := parseOrEmpty(raw)
	if err != nil {
		return nil, err
	}
	bar.Set(key, value)
	updated, err := bar.Indented()
	if err != nil {
		return nil, err
	}
	document.Set("bar", updated)
	return document.Indented()
}

func unmarked(data []byte) []byte {
	return bytes.TrimPrefix(data, []byte("\ufeff"))
}

func parseOrEmpty(data []byte) (*ordered.Object, error) {
	text := bytes.TrimSpace(data)
	if len(text) == 0 || bytes.Equal(text, []byte("null")) {
		return &ordered.Object{}, nil
	}
	return ordered.Parse(data)
}

func SetBar(path, key string, value json.RawMessage) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o777); err != nil {
		return err
	}
	unlock, err := hold(filepath.Dir(path))
	if err != nil {
		return err
	}
	defer unlock()
	data, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	updated, err := WithBar(data, key, value)
	if err != nil {
		return fmt.Errorf("%s is not valid JSON; fix it first", path)
	}
	if err := atomic.Write(path, updated, 0o600); err != nil {
		return err
	}
	return os.Chmod(path, 0o600)
}
