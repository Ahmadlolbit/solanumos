package engine

import (
	"bytes"
	"errors"
	"strings"

	"github.com/Ahmadlolbit/solanumos/internal/move"
	"github.com/Ahmadlolbit/solanumos/internal/run"
	"github.com/Ahmadlolbit/solanumos/internal/ui"
)

var errNotMoved = errors.New("this computer still has Hyprtrixie's folders: run sudo solanum install first, which moves it over to Solanum")

type carried struct {
	output bytes.Buffer
	notes  []string
}

func carryOver(mode string, options Options) (*carried, error) {
	moving := &carried{}
	if !move.Pending() {
		move.Hold()
		return moving, nil
	}
	if mode != installMode && !strings.HasPrefix(mode, appMode) {
		return nil, errNotMoved
	}
	if !options.Events {
		ui.Busy("Move this computer from Hyprtrixie to Solanum")
	}
	moving.notes = move.Run(&moving.output)
	if !options.Events {
		ui.Done()
	}
	return moving, nil
}

func (work *session) carried(moving *carried) {
	if moving.output.Len() == 0 {
		return
	}
	work.log.Write(moving.output.Bytes())
	for _, note := range moving.notes {
		work.log.Line("info", note)
		ui.Note(note)
	}
}

func (work *session) settle() {
	if !work.changesFiles() || work.runner.Interrupted || !move.Due() {
		run.Release()
		return
	}
	if !work.options.Events {
		ui.Busy("Rebuild the boot image and the boot menu once")
	}
	err := move.Settle(work.log)
	if !work.options.Events {
		ui.Done()
	}
	if err != nil {
		work.log.Line("warn", "the boot image and the boot menu weren't rebuilt, so the next run does it: "+err.Error())
		ui.Warn("The boot image and the boot menu weren't rebuilt; the next sudo solanum install does it")
	}
}
