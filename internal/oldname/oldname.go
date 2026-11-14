package oldname

import (
	"os"
	"strings"
)

const (
	Name          = "hyprtrixie"
	VersionKey    = "HYPRTRIXIE_CONFIG_VERSION"
	Lock          = "/run/hyprtrixie.lock"
	AddressKey    = "X-Hyprtrixie-Address="
	FirstSnapshot = "hyprtrixie: before setup"
	OldSuffix     = ".hyprtrixie-old"
	NewSuffix     = ".hyprtrixie-new"
	PartSuffix    = ".hyprtrixie-part"
	VersionStamp  = ".hyprtrixie-version"
	KDEBackup     = ".before-hyprtrixie"
	KDEScheme     = "ColorScheme=Hyprtrixie"
	KDERecord     = "[Hyprtrixie]"
	Prefix        = "hyprtrixie-"
	EditorPrefix  = "Hyprtrixie "
	Extension     = "hyprtrixie.themes"
	FirewallTag   = "hyprtrixie "
	LogStart      = "hyprtrixie "
	Policy        = "hyprtrixie-theme.json"
	FirefoxStyle  = "hyprtrixie.css"
	FirefoxImport = `@import url("hyprtrixie.css");`
	Unfinished    = "hyprtrixie-unfinished"
	ThemesSource  = "hyprtrixie-themes"
	UpdatePart    = "hyprtrixie"
	Clone         = "hyprtrixie"
)

var (
	forward  = strings.NewReplacer("hyprtrixie", "solanum", "Hyprtrixie", "Solanum", "HYPRTRIXIE", "SOLANUM")
	backward = strings.NewReplacer("solanum", "hyprtrixie", "Solanum", "Hyprtrixie", "SOLANUM", "HYPRTRIXIE")
)

func Swap(text string) string {
	return forward.Replace(text)
}

func Back(text string) string {
	return backward.Replace(text)
}

func Prefer(path string) string {
	if _, err := os.Lstat(path); err == nil {
		return path
	}
	if old := Back(path); old != path {
		if _, err := os.Stat(old); err == nil {
			return old
		}
	}
	return path
}
