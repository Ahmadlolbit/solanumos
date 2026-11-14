package main

import (
	"errors"
	"os"

	"github.com/Ahmadlolbit/solanumos/internal/account"
	"github.com/Ahmadlolbit/solanumos/internal/engine"
	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/ui"
)

func handedOver(mode string, body func([]string) int) int {
	if os.Getenv("SOLANUM_ENGINE") == "" {
		return throughScript()
	}
	args := os.Args[2:]
	_, err := engine.Parse(args)
	switch {
	case errors.Is(err, engine.ErrHelp):
		return showHelp()
	case err != nil:
		return fail("%v", err)
	}
	return asRoot(mode, func() int { return body(args) })
}

func testingLogs() {
	person, ok := account.Target()
	if !ok {
		ui.Warn("no testing logs: cannot tell whose desktop this is")
		return
	}
	ui.Busy("Pack the testing logs")
	saved, err := gather(person, env.Root(), "solanum-logs-", true)
	ui.Done()
	if err != nil {
		ui.Warn(err.Error())
		return
	}
	ui.Print(saved)
}

func installCommand() int {
	engine.AfterTesting = testingLogs
	return handedOver("install", engine.Install)
}

func accountCommand() int {
	return handedOver("account", engine.Account)
}

func verifyCommand() int {
	return handedOver("verify", engine.Verify)
}
