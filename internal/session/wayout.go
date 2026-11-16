package session

import (
	"time"

	"github.com/Ahmadlolbit/solanumos/internal/bootimage"
	"github.com/Ahmadlolbit/solanumos/internal/run"
	"github.com/Ahmadlolbit/solanumos/internal/snapshots"
)

const (
	updatePoint = "before solanum update"
	loginScreen = "Ctrl+Alt+F1 brings back the login screen."
	sendLogs    = "solanum collect packs the logs to send with a bug report."
	likelyCause = "The graphics driver or the kernel is the likely cause, not your configuration."
)

type WayOut struct {
	Running  string
	Previous string
	Restore  snapshots.Snapshot
}

func FindWayOut(limit time.Duration) WayOut {
	var way WayOut
	kernels := bootimage.Kernels()
	running, _ := run.CaptureQuiet(limit, "uname", "-r")
	if count := len(kernels); count > 1 && kernels[count-1] == running {
		way.Running, way.Previous = running, kernels[count-2]
	}
	points, _ := snapshots.All()
	for _, point := range points {
		if point.Description == updatePoint {
			way.Restore = point
		}
	}
	return way
}

func (way WayOut) restoreAdvice(afterKernel bool) string {
	if way.Restore.Number == "" {
		return "solanum restore, without sudo, lists the restore points to go back to."
	}
	text := "Go back to before the last update"
	if afterKernel {
		text = "If that doesn't help, go back to before the last update"
	}
	return text + ": solanum restore " + way.Restore.Number + ", without sudo (" + way.Restore.Label() + ")."
}

func (ending Ending) Advice(way WayOut) []string {
	var advice []string
	if ending.Graphics {
		cause := likelyCause
		if way.Previous != "" {
			cause += " This boot runs the newest kernel, " + way.Running + ": restart, press Esc right after the laptop's logo, open Advanced options and pick " + way.Previous + "."
		}
		advice = append(advice, cause)
	}
	advice = append(advice, way.restoreAdvice(ending.Graphics && way.Previous != ""))
	return append(advice, loginScreen, sendLogs)
}
