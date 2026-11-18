package session

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Ahmadlolbit/solanumos/internal/account"
	"github.com/Ahmadlolbit/solanumos/internal/atomic"
	"github.com/Ahmadlolbit/solanumos/internal/fsutil"
)

const (
	LastLog   = OwnLog + ".1"
	ToldNote  = ".local/state/solanum/desktop-crash-told"
	lastLines = 30
)

func modified(path string) time.Time {
	info, err := os.Lstat(path)
	if err != nil {
		return time.Time{}
	}
	return info.ModTime()
}

func CameUpLastTime(home string, uid int) (string, time.Time, bool) {
	path := filepath.Join(home, LastLog)
	began, ended, up := logStart(path, uid), modified(path), modified(filepath.Join(home, GoodLogin))
	if began.IsZero() || up.Before(began) || up.After(ended.Add(slack)) {
		return "", time.Time{}, false
	}
	return began.Format(time.DateTime), up.Truncate(time.Second), true
}

func LastBound(home string, uid int) time.Time {
	if _, since, up := CameUpLastTime(home, uid); up {
		return since
	}
	if began := logStart(filepath.Join(home, LastLog), uid); !began.IsZero() {
		return began.Add(-slack)
	}
	return time.Time{}
}

func LastTime(person account.Account, since time.Time, limit time.Duration) Ending {
	var ending Ending
	uid, err := strconv.Atoi(person.UID)
	if err != nil || !filepath.IsAbs(person.Home) || since.IsZero() {
		return ending
	}
	ending.fromDump(person.UID, since, limit)
	ending.fromReport(person.Home, uid, since)
	return ending
}

func LastLines(home string, uid int) []string {
	return tailOf(filepath.Join(home, LastLog), uid, lastLines)
}

func Told(home string) string {
	return strings.TrimSpace(fsutil.Optional(filepath.Join(home, ToldNote)))
}

func NoteTold(home, key string) error {
	path := filepath.Join(home, ToldNote)
	if err := os.MkdirAll(filepath.Dir(path), ownMode); err != nil {
		return err
	}
	return atomic.Write(path, []byte(key+"\n"), logMode)
}
