package restic

import (
	"cmp"
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"github.com/Ahmadlolbit/solanumos/internal/fsutil"
)

const (
	StaleAfter = 7 * 24 * time.Hour
	dueAfter   = 20 * time.Hour
)

type Backup struct {
	ID    string   `json:"id"`
	Time  string   `json:"time"`
	Paths []string `json:"paths"`
}

func DailyService(program string) string {
	return fmt.Sprintf(`[Unit]
Description=Back up your home folder once a day, on the charger

[Service]
Type=oneshot
Nice=10
IOSchedulingClass=idle
ExecStart="%s" scheduled
`, execEscape.Replace(program))
}

func DailyTimer() string {
	return `[Unit]
Description=Look every hour whether the day's backup is due

[Timer]
OnCalendar=hourly
Persistent=true
RandomizedDelaySec=10min

[Install]
WantedBy=timers.target
`
}

func Due(last, now time.Time) bool {
	return last.IsZero() || now.Sub(last) >= dueAfter
}

func Stale(last, since, now time.Time) bool {
	from := last
	if from.IsZero() {
		from = since
	}
	return !from.IsZero() && now.Sub(from) >= StaleAfter
}

func Reachable(location string) bool {
	_, path, ok := local(location)
	return !ok || fsutil.IsDir(path)
}

type listed struct {
	Time    time.Time `json:"time"`
	ShortID string    `json:"short_id"`
	ID      string    `json:"id"`
	Paths   []string  `json:"paths"`
}

func Backups(data []byte) ([]Backup, error) {
	var found []listed
	if err := json.Unmarshal(data, &found); err != nil {
		return nil, err
	}
	slices.SortStableFunc(found, func(first, second listed) int {
		return second.Time.Compare(first.Time)
	})
	backups := []Backup{}
	for _, each := range found {
		paths := each.Paths
		if paths == nil {
			paths = []string{}
		}
		backups = append(backups, Backup{cmp.Or(each.ShortID, each.ID), each.Time.Format(time.RFC3339), paths})
	}
	return backups, nil
}
