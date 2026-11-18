package hypr

import (
	"bytes"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/Ahmadlolbit/solanumos/internal/account"
	"github.com/Ahmadlolbit/solanumos/internal/run"
)

const (
	verdictLine  = "Config parsing result:"
	verdictShown = 20
)

func isVerdict(line string) bool {
	return strings.Contains(line, verdictLine)
}

func lastLines(lines []string) string {
	shown := lines[max(0, len(lines)-verdictShown):]
	return strings.Join(shown, "\n") + "\n"
}

func Verdict(lines []string) (bool, string) {
	start := slices.IndexFunc(lines, isVerdict)
	if start < 0 {
		return false, ""
	}
	verdict := slices.IndexFunc(lines[start:], func(line string) bool {
		return !isVerdict(line) && strings.Trim(line, " \t") != ""
	})
	if verdict >= 0 && lines[start+verdict] == "config ok" {
		return true, ""
	}
	shown := lines[start:]
	return false, strings.Join(shown[:min(verdictShown, len(shown))], "\n") + "\n"
}

func verify(log io.Writer, person account.Account, runtime string, output io.Writer) error {
	args := []string{"env", "XDG_RUNTIME_DIR=" + runtime, "Hyprland", "--verify-config"}
	if person.Name == "" {
		return run.LoggedWithin(run.Minute, output, args[0], args[1:]...)
	}
	if err := run.Logged(log, "chown", person.Name+":", runtime); err != nil {
		return err
	}
	return person.RunWithin(run.Minute, output, args...)
}

func Accepts(log io.Writer, person account.Account) (bool, string) {
	runtime, err := os.MkdirTemp("", "solanum")
	if err != nil {
		return false, err.Error() + "\n"
	}
	defer os.RemoveAll(runtime)
	var output bytes.Buffer
	err = verify(log, person, runtime, &output)
	lines := strings.Split(strings.TrimRight(output.String(), "\n"), "\n")
	if ok, complaint := Verdict(lines); ok || complaint != "" {
		return ok, complaint
	}
	switch {
	case strings.TrimSpace(output.String()) != "":
		return false, lastLines(lines)
	case err != nil:
		return false, err.Error() + "\n"
	}
	return false, ""
}
