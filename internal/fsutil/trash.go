package fsutil

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Ahmadlolbit/solanumos/internal/atomic"
)

func Trash(path, trash string) error {
	files, info := filepath.Join(trash, "files"), filepath.Join(trash, "info")
	absolute, err := filepath.Abs(path)
	if err == nil {
		err = os.MkdirAll(files, 0o700)
	}
	if err == nil {
		err = os.MkdirAll(info, 0o700)
	}
	if err != nil {
		return err
	}
	note, err := atomic.Fresh(info, filepath.Base(path), ".trashinfo", 0o600)
	if err != nil {
		return err
	}
	text := "[Trash Info]\nPath=" + (&url.URL{Path: absolute}).EscapedPath() + "\nDeletionDate=" + time.Now().Format("2006-01-02T15:04:05") + "\n"
	err = atomic.Write(note, []byte(text), 0o600)
	if err == nil {
		err = atomic.Move(path, filepath.Join(files, strings.TrimSuffix(filepath.Base(note), ".trashinfo")))
	}
	if err != nil {
		os.Remove(note)
	}
	return err
}
