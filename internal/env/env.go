package env

import (
	"os"
	"os/user"
	"path/filepath"
)

const Version = "0.9.0"

func Or(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func Arg(index int) string {
	if index < len(os.Args) {
		return os.Args[index]
	}
	return ""
}

func Home() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if current, err := user.Current(); err == nil && current.HomeDir != "" {
		return current.HomeDir
	}
	return "/"
}

func ConfigHome() string {
	return Or("XDG_CONFIG_HOME", filepath.Join(Home(), ".config"))
}

func StateHome() string {
	return Or("XDG_STATE_HOME", filepath.Join(Home(), ".local", "state"))
}

func ThemeRoot() string {
	return Or("SOLANUM_THEME_ROOT", ShareDir())
}

func AppConfig() string {
	return filepath.Join(ConfigHome(), "hyprtrixie")
}

func LocalState() string {
	return filepath.Join(Home(), ".local", "state")
}

func ThemeState() string {
	return filepath.Join(LocalState(), "hyprtrixie", "current")
}

func CacheHome() string {
	return Or("XDG_CACHE_HOME", filepath.Join(Home(), ".cache"))
}

func DataHome() string {
	return Or("XDG_DATA_HOME", filepath.Join(Home(), ".local/share"))
}

func ShareDir() string {
	return Or("SOLANUM_SHARE_DIR", "/usr/share/hyprtrixie")
}

func Root() string {
	return os.Getenv("SOLANUM_ROOT")
}

func SystemClone() string {
	return Or("SOLANUM_SYSTEM_CLONE", "/opt/hyprtrixie")
}

func BinDir() string {
	return Or("SOLANUM_BIN_DIR", "/usr/local/bin")
}

func ToolsDir() string {
	return Or("SOLANUM_TOOLS_DIR", filepath.Join(Root(), "build", "bin"))
}

func Manifests() string {
	return Or("SOLANUM_MANIFESTS", filepath.Join(Or("SOLANUM_ROOT", "."), "manifests"))
}

func CacheDir() string {
	return Or("SOLANUM_CACHE_DIR", "/var/cache/hyprtrixie")
}

func LogDir() string {
	return Or("SOLANUM_LOG_DIR", "/var/log/hyprtrixie")
}

func StateDir() string {
	return Or("SOLANUM_STATE_DIR", "/var/lib/hyprtrixie")
}
