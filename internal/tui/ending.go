package tui

import (
	"fmt"
	"os"
	"slices"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Ahmadlolbit/solanumos/internal/disks"
	"github.com/Ahmadlolbit/solanumos/internal/luks"
	"github.com/Ahmadlolbit/solanumos/internal/osinstall"
	"github.com/Ahmadlolbit/solanumos/internal/qr"
)

const keptTitle = "Leave before you've kept the recovery key?"

var codeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Background(lipgloss.Color("0"))

type ending struct {
	viewer
	flow *wizard
	key  string
	kept bool
	hand *page
}

func newEnding(flow *wizard) *ending {
	end := &ending{viewer: viewer{leave: func() tea.Cmd { return tea.Quit }}, flow: flow}
	end.hand = &page{wizard: flow, title: "Recovery key", rows: end.handRows}
	return end
}

func (end *ending) handing() bool {
	return end.key != "" && !end.kept
}

func (end *ending) name() string {
	if end.handing() {
		return "Recovery key"
	}
	return "Summary"
}

func (end *ending) arrive() {
	if end.handing() {
		end.hand.arrive()
	}
}

func (end *ending) press(key tea.KeyMsg) tea.Cmd {
	if !end.handing() {
		_, cmd := end.viewer.press(key)
		return cmd
	}
	if key.String() == "esc" && end.hand.search == nil && end.hand.filter == "" {
		return nil
	}
	return end.hand.press(key)
}

func (end *ending) draw(width, height int) string {
	if end.handing() {
		return end.hand.draw(width, height)
	}
	return end.viewer.draw(width, height)
}

func (end *ending) help() string {
	if end.handing() {
		return end.hand.help()
	}
	return ""
}

func (end *ending) keys() string {
	if end.handing() {
		return "Arrows move · Enter picks"
	}
	return end.viewer.keys()
}

func (end *ending) typing() bool {
	return false
}

func (end *ending) leaving() (string, func() tea.Cmd) {
	if end.handing() {
		return keptTitle, end.leave
	}
	return "", end.leave
}

func (end *ending) handRows() []row {
	rows := []row{{label: "Your recovery key", heading: true}, {info: true}}
	for _, line := range luks.Lines(end.key) {
		rows = append(rows, row{label: line, info: true})
	}
	return append(rows, row{info: true},
		row{label: "It opens the disk when the disk password is forgotten, or when the TPM refuses after a firmware change. Type it in small letters, with its dashes, wherever the disk password is asked for.", info: true},
		row{label: "This is the only time it's shown. Keep it off this computer: on paper, or on your phone.", info: true},
		row{info: true},
		row{label: "Show the QR code", help: "A phone that scans it keeps the key", enter: end.showCode},
		row{label: "Save to a USB stick", help: "A text file with the key, on a USB stick other than the installer's", enter: end.saveKey},
		row{label: "I've kept it", help: "On to the summary", enter: end.keep},
	)
}

func (end *ending) keep() tea.Cmd {
	end.kept = true
	return nil
}

func codeColours(lines []string) []string {
	shown := make([]string, len(lines))
	for index, line := range lines {
		shown[index] = codeStyle.Render(line)
	}
	return shown
}

func (end *ending) showCode() tea.Cmd {
	lines, err := qr.Lines(end.key, os.Getenv("TERM") == "linux")
	if err != nil {
		end.flow.notice = "The QR code can't be drawn: " + err.Error()
		return nil
	}
	width, height, _, _ := end.flow.frame()
	if len(lines) > height || lipgloss.Width(lines[0]) > width {
		end.flow.notice = fmt.Sprintf("The QR code needs %d by %d and the screen has %d by %d: make it bigger, or save the key to a USB stick", lipgloss.Width(lines[0]), len(lines), width, height)
		return nil
	}
	return end.flow.open(&viewer{lines: codeColours(lines), leave: func() tea.Cmd { return nil }})
}

func (end *ending) host() (string, string) {
	if end.flow.system == nil {
		return "", "solanum"
	}
	return end.flow.system.answers.Disk, end.flow.system.answers.Host
}

func (end *ending) saveKey() tea.Cmd {
	target, _ := end.host()
	found, err := osinstall.KeySticks(target)
	switch {
	case err != nil:
		end.flow.notice = "Couldn't look for USB sticks: " + err.Error()
		return nil
	case len(found) == 0:
		end.flow.notice = "Put in a USB stick other than the installer's, then choose Save to a USB stick again"
		return nil
	case len(found) == 1:
		return end.saveTo(found[0])
	}
	names := make([]string, len(found))
	for index, stick := range found {
		names[index] = stick.Name()
	}
	return end.flow.open(question("Which USB stick?", names, func(text string, ok bool) tea.Cmd {
		if at := slices.Index(names, text); ok && at >= 0 {
			return end.saveTo(found[at])
		}
		return nil
	}))
}

func (end *ending) saveTo(stick disks.Stick) tea.Cmd {
	_, host := end.host()
	name, err := osinstall.SaveKey(stick, end.key, host)
	if err != nil {
		end.flow.notice = "The key wasn't saved: " + err.Error()
		return nil
	}
	end.flow.notice = "Saved to " + stick.Name() + " as " + name + ". Open it on another computer before you count on it"
	return nil
}
