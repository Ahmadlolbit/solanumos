package notify

import (
	"slices"
	"strings"
	"time"

	"github.com/Ahmadlolbit/solanumos/internal/run"
)

const (
	limit   = 30 * time.Second
	service = "org.freedesktop.Notifications"
	between = time.Second / 4
)

var markup = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")

func Markup(text string) string {
	return markup.Replace(text)
}

func arguments(options []string, summary, body string) []string {
	args := append(slices.Clone(options), "--", summary)
	if body != "" {
		args = append(args, Markup(body))
	}
	return args
}

func Post(options []string, summary, body string) int {
	return run.Within(limit, "notify-send", arguments(options, summary, body)...)
}

func Ask(options []string, summary, body string) string {
	output, _ := run.Output("notify-send", arguments(options, summary, body)...)
	return strings.TrimSpace(output)
}

func Ready(wait time.Duration) bool {
	for started := time.Now(); ; time.Sleep(between) {
		if _, err := run.CaptureQuiet(run.Quick, "busctl", "--user", "status", service); err == nil {
			return true
		}
		if time.Since(started) >= wait {
			return false
		}
	}
}
