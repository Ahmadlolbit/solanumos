package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/luks"
	"github.com/Ahmadlolbit/solanumos/internal/qr"
	"github.com/Ahmadlolbit/solanumos/internal/run"
	"github.com/Ahmadlolbit/solanumos/internal/term"
	"github.com/Ahmadlolbit/solanumos/internal/ui"
	"github.com/Ahmadlolbit/solanumos/internal/words"
)

const (
	wrongPassword = 2
	keyUsage      = "usage: solanum recovery-key status|new|remove"
)

func keyRoot() rootWords {
	return rootWords{
		words.T("the root isn't on an encrypted disk, so there's no disk for a recovery key to open"),
		words.T("the root is on more than one disk, and solanum recovery-key handles a root on one"),
		words.T("%s is LUKS1, which can't hold a recovery key"),
	}
}

func hasRecoveryKey(device string) (bool, error) {
	dump, err := run.CaptureQuiet(run.Quick, "cryptsetup", "luksDump", device)
	if err != nil {
		return false, fmt.Errorf(words.T("%s's key slots can't be read"), device)
	}
	return luks.HasKey(dump), nil
}

func keyStatus() int {
	root, err := luksRootFor(keyRoot())
	if err != nil {
		return say(sentenceOf(err.Error()) + ".\n")
	}
	has, err := hasRecoveryKey(root.device)
	switch {
	case err != nil:
		return fail("%s.", sentenceOf(err.Error()))
	case has:
		return say(words.T("This disk has a recovery key.") + "\n")
	}
	return say(words.T("This disk has no recovery key. Make one: sudo solanum recovery-key new") + "\n")
}

func sentenceOf(text string) string {
	if text == "" {
		return text
	}
	return strings.ToUpper(text[:1]) + text[1:]
}

func checkPassword(device, password string) error {
	_, messages, err := run.CaptureFed(run.Minute, password, "cryptsetup", "open", "--test-passphrase", "--key-file=-", device)
	switch {
	case err == nil:
		return nil
	case run.Code(err) == wrongPassword:
		return errors.New(words.T("that isn't the disk password"))
	}
	return fmt.Errorf(words.T("the password couldn't be checked: %s"), luks.Reason(messages))
}

func askToReplace(device string) error {
	has, err := hasRecoveryKey(device)
	switch {
	case err != nil:
		return err
	case has && !ui.Confirm(words.T("The recovery key you have stops working. Make a new one?")):
		return errors.New(words.T("nothing was changed"))
	}
	return nil
}

func showKey(key string) {
	ui.Print("\n" + words.T("Your recovery key:") + "\n\n")
	for _, line := range luks.Lines(key) {
		ui.Print("  " + line + "\n")
	}
	if code, err := qr.Lines(key, os.Getenv("TERM") == "linux"); err == nil {
		ui.Print("\n" + paintedCode(code) + "\n")
	}
	ui.Print("\n" + words.T("Type it in small letters, with its dashes. Keep it off this computer: on paper, or on your phone.") + "\n")
}

func paintedCode(lines []string) string {
	if !term.IsTerminal(os.Stdout) || os.Getenv("NO_COLOR") != "" {
		return strings.Join(lines, "\n")
	}
	for index, line := range lines {
		lines[index] = "\x1b[97;40m" + line + "\x1b[0m"
	}
	return strings.Join(lines, "\n")
}

func newKey() int {
	root, err := luksRootFor(keyRoot())
	if err != nil {
		return fail("%s.", sentenceOf(err.Error()))
	}
	if !ui.CanAsk() {
		return fail("%s", words.T("Run it in a terminal: it asks for your disk password, and shows the key."))
	}
	ui.Print(words.T("A recovery key opens the disk when the disk password is forgotten, or when the TPM refuses after a firmware change. It works wherever the disk password is asked for.") + "\n")
	if err := askToReplace(root.device); err != nil {
		return fail("%s.", sentenceOf(err.Error()))
	}
	password, ok := ui.Password(words.T("Disk password:"))
	if !ok {
		return fail("%s", words.T("Nothing was changed."))
	}
	if err := checkPassword(root.device, password); err != nil {
		return fail(words.T("%s. Nothing was changed."), sentenceOf(err.Error()))
	}
	key, _, err := luks.Make(root.device, password)
	if key == "" {
		return fail(words.T("The recovery key wasn't made: %v. The disk opens as before."), err)
	}
	showKey(key)
	if err != nil {
		return fail(words.T("The old recovery key may still work too: %v."), err)
	}
	return ui.ExitCode()
}

func removeKey() int {
	root, err := luksRootFor(keyRoot())
	if err != nil {
		return fail("%s.", sentenceOf(err.Error()))
	}
	has, err := hasRecoveryKey(root.device)
	switch {
	case err != nil:
		return fail("%s.", sentenceOf(err.Error()))
	case !has:
		return say(words.T("This disk has no recovery key.") + "\n")
	case !ui.Confirm(words.T("Take the recovery key away? It stops opening the disk.")):
		return fail("%s", words.T("Nothing was changed."))
	}
	args := luks.Wipe(root.device)
	if _, messages, err := run.CaptureFed(run.Minute, "", args[0], args[1:]...); err != nil {
		return fail(words.T("The recovery key is still there: %s failed. %s"), args[0], luks.Reason(messages))
	}
	return say(words.T("Done. The recovery key no longer opens the disk.") + "\n")
}

func recoveryKeyCommand() int {
	commands := map[string]func() int{"status": keyStatus, "": keyStatus, "new": alone(newKey), "remove": alone(removeKey)}
	command, known := commands[env.Arg(2)]
	if !known || env.Arg(3) != "" {
		return fail(keyUsage)
	}
	return asRoot(strings.TrimSpace("recovery-key "+env.Arg(2)), command)
}
