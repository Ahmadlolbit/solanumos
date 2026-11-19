package inhibit

import (
	"cmp"
	"sync/atomic"
	"time"

	"github.com/Ahmadlolbit/solanumos/internal/run"
	"github.com/Ahmadlolbit/solanumos/internal/ui"
)

const (
	what   = "--what=sleep:shutdown:handle-lid-switch"
	settle = 500 * time.Millisecond
)

func While(why string) func() {
	child, err := run.SpawnWith(run.Spawning{Apart: true}, "systemd-inhibit", what, "--who=Solanum", "--why="+why, "--mode=block", "cat")
	if err != nil {
		return func() {}
	}
	var released atomic.Bool
	ended := make(chan struct{})
	go func() {
		_, said := child.Wait()
		if !released.Load() {
			ui.Print("The computer may sleep or shut down during this, as systemd-inhibit couldn't hold them off: " + cmp.Or(said, "it ended early") + "\n")
		}
		close(ended)
	}()
	select {
	case <-ended:
	case <-time.After(settle):
	}
	return func() {
		released.Store(true)
		child.Input.Close()
		<-ended
	}
}
