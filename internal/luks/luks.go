package luks

import (
	"errors"
	"regexp"
	"slices"
	"strings"

	"github.com/Ahmadlolbit/solanumos/internal/run"
)

const (
	Characters = "cbdefghijklnrtuv"
	Enroller   = "systemd-cryptenroll"
	token      = "systemd-recovery"
	group      = 8
	perLine    = 4
	hidden     = "(hidden)"
)

var (
	shape     = regexp.MustCompile(`^[` + Characters + `]{8}(-[` + Characters + `]{8}){7}$`)
	anywhere  = regexp.MustCompile(`[` + Characters + `]{8}(-[ \t]*(\r?\n[ \t]*)?[` + Characters + `]{8}){7}|[` + Characters + `]{64}`)
	modhexOne = func() (table [256]bool) {
		for _, character := range Characters {
			table[character] = true
		}
		return table
	}()
)

func Enroll(device string) []string {
	return []string{Enroller, "--unlock-key-file=/dev/stdin", "--recovery-key", "--wipe-slot=recovery", device}
}

func Wipe(device string) []string {
	return []string{Enroller, "--wipe-slot=recovery", device}
}

func MakesKey(args []string) bool {
	return len(args) > 0 && args[0] == Enroller && slices.Contains(args, "--recovery-key")
}

func Valid(key string) bool {
	return shape.MatchString(key)
}

func Key(output string) (string, bool) {
	for _, word := range strings.Fields(output) {
		if Valid(word) {
			return word, true
		}
	}
	return "", false
}

func Lines(key string) []string {
	if !Valid(key) {
		return []string{key}
	}
	groups := strings.Split(key, "-")
	return []string{strings.Join(groups[:perLine], "-") + "-", strings.Join(groups[perLine:], "-")}
}

func Mentioned(data []byte) bool {
	run := 0
	for _, character := range data {
		switch {
		case modhexOne[character]:
			run++
			if run == 8*group {
				return true
			}
		case character == '-' && run >= group:
			return true
		default:
			run = 0
		}
	}
	return false
}

func Hide(text string) string {
	if !Mentioned([]byte(text)) {
		return text
	}
	return anywhere.ReplaceAllString(text, hidden)
}

func Reason(messages string) string {
	lines := strings.Split(strings.TrimSpace(Hide(messages)), "\n")
	return strings.TrimSpace(lines[len(lines)-1])
}

func Make(device, password string) (string, string, error) {
	args := Enroll(device)
	output, messages, err := run.CaptureFed(run.Minute, password, args[0], args[1:]...)
	key, found := Key(output)
	switch {
	case err != nil && Reason(messages) != "":
		return "", Hide(messages), errors.New(Reason(messages))
	case err != nil:
		return "", Hide(messages), err
	case !found:
		return "", Hide(messages), errors.New(Enroller + " showed no recovery key")
	}
	return key, Hide(messages), nil
}

func section(line string) (string, bool) {
	if line == "" || strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
		return "", false
	}
	return strings.TrimSpace(line), true
}

func tokenKind(line string) (string, bool) {
	number, kind, found := strings.Cut(strings.TrimSpace(line), ": ")
	if !found || number == "" || strings.Trim(number, "0123456789") != "" {
		return "", false
	}
	fields := strings.Fields(kind)
	if len(fields) == 0 {
		return "", false
	}
	return fields[0], true
}

func HasKey(dump string) bool {
	inTokens, recovery := false, false
	for _, line := range strings.Split(dump, "\n") {
		if name, found := section(line); found {
			inTokens, recovery = name == "Tokens:", false
			continue
		}
		if !inTokens {
			continue
		}
		if kind, found := tokenKind(line); found {
			recovery = kind == token
			continue
		}
		if recovery && strings.HasPrefix(strings.TrimSpace(line), "Keyslot:") {
			return true
		}
	}
	return false
}
