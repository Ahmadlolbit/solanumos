package settings

const (
	Toggle = "toggle"
	Pick   = "choice"
	Slide  = "slider"
	Time   = "time"
	Text   = "text"
	Button = "button"

	Yours    = "yours"
	Computer = "computer"

	shellPath = "/usr/share/solanum/shell"
)

type Page struct {
	ID       string   `json:"id"`
	Label    string   `json:"label"`
	Glyph    string   `json:"glyph"`
	Keywords []string `json:"keywords,omitempty"`
}

type Choice struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type Row struct {
	Name     string   `json:"name"`
	Label    string   `json:"label"`
	Page     string   `json:"page"`
	Kind     string   `json:"kind"`
	Value    string   `json:"value"`
	Choices  []Choice `json:"choices,omitempty"`
	Multiple bool     `json:"multiple,omitempty"`
	Min      float64  `json:"min,omitempty"`
	Max      float64  `json:"max,omitempty"`
	Step     float64  `json:"step,omitempty"`
	Whose    string   `json:"whose"`
	Twin     string   `json:"twin,omitempty"`
	After    string   `json:"after,omitempty"`
	Help     string   `json:"help,omitempty"`
	Keywords []string `json:"keywords,omitempty"`
	Command  []string `json:"command,omitempty"`
	Hidden   bool     `json:"hidden,omitempty"`
}

type Listing struct {
	Pages []Page `json:"pages"`
	Rows  []Row  `json:"rows"`
}

var pages = []Page{
	{"about", "About", "\uf05a", []string{"version", "system", "hardware", "credits", "licences", "debian"}},
	{"themes", "Appearance and themes", "\uf53f", []string{"theme", "colours", "colors", "accent", "dark", "light", "fonts", "style", "your own theme"}},
	{"wallpaper", "Wallpaper and desktop", "\uf03e", []string{"wallpaper", "background", "pictures", "slideshow", "logo"}},
	{"display", "Display", "\uf108", []string{"screen", "scale", "brightness", "night light", "corners"}},
	{"sound", "Sound", "\uf028", []string{"audio", "volume", "speakers", "microphone", "chime"}},
	{"network", "Network", "\uf1eb", []string{"wi-fi", "wifi", "internet", "networkmanager"}},
	{"bluetooth", "Bluetooth", "\uf293", []string{"devices", "pair", "headphones"}},
	{"power", "Power and battery", "\uf240", []string{"battery", "power profile", "suspend", "hibernate", "charge"}},
	{"lock", "Lock screen and idle", "\uf023", []string{"lock", "idle", "screensaver", "login screen"}},
	{"keyboard", "Keyboard", "\uf11c", []string{"layout", "keys", "shortcuts", "bindings"}},
	{"mouse", "Mouse and touchpad", "\uf245", []string{"touchpad", "trackpoint", "scrolling", "tap"}},
	{"bar", "Bar, dock and menu", "\uf0c9", []string{"bar", "widgets", "clock", "weather", "zen"}},
	{"notifications", "Notifications", "\uf0f3", []string{"do not disturb", "quiet hours", "tips", "warnings"}},
	{"windows", "Windows and workspaces", "\uf2d2", []string{"windows", "animations", "workspaces"}},
	{"apps", "Apps and default apps", "\uf009", []string{"apps", "applications", "flatpak", "extras", "profile"}},
	{"privacy", "Privacy and security", "\uf132", []string{"firewall", "security"}},
	{"accounts", "Accounts and family", "\uf007", []string{"fingerprint", "account", "user"}},
	{"sharing", "Sharing and remote access", "\uf1e0", []string{"ssh", "remote"}},
	{"time", "Date and time", "\uf017", []string{"clock", "24-hour", "12-hour", "location", "sun"}},
	{"accessibility", "Accessibility", "\uf29a", []string{"screen reader", "orca", "text size", "larger text"}},
	{"updates", "Updates, snapshots and backups", "\uf021", []string{"updates", "snapshots", "restore", "mirror"}},
	{"storage", "Storage and memory", "\uf1c0", []string{"swap", "zram", "memory", "disk"}},
	{"startup", "Startup, login and boot", "\uf011", []string{"boot", "splash", "grub", "console"}},
	{"hardware", "Hardware and drivers", "\uf2db", []string{"kernel", "graphics", "nvidia", "webcam", "drivers"}},
	{"developer", "Developer", "\uf121", []string{"docker", "podman", "mise", "prompt", "containers"}},
}

func Pages() []Page {
	return pages
}

func KnownPage(id string) bool {
	for _, page := range pages {
		if page.ID == id {
			return true
		}
	}
	return false
}

func popup(name string) []string {
	return []string{"qs", "-p", shellPath, "ipc", "call", "popup", "toggle", name}
}

func buttons() []Row {
	return []Row{
		{Name: "open.wifi", Label: "Wi-Fi networks", Page: "network", Help: "Join a network, see the ones in range and trust the one you're on", Keywords: []string{"wi-fi", "wifi", "join"}, Command: popup("wifi")},
		{Name: "open.bluetooth", Label: "Bluetooth devices", Page: "bluetooth", Help: "Turn Bluetooth on, pair, connect and see a device's battery", Keywords: []string{"pair", "connect"}, Command: popup("bluetooth")},
		{Name: "open.audio", Label: "Output, input and volume", Page: "sound", Help: "Speakers, microphones and the volume of each app", Keywords: []string{"audio", "volume", "speakers", "microphone"}, Command: popup("audio")},
		{Name: "open.brightness", Label: "Brightness", Page: "display", Help: "The screen's brightness and the keyboard's light", Keywords: []string{"backlight"}, Command: popup("brightness")},
		{Name: "open.battery", Label: "Power mode, battery and charge limit", Page: "power", Help: "The power profile, the battery's health and where charging stops", Keywords: []string{"battery", "charge limit", "power profile"}, Command: popup("battery")},
		{Name: "open.layout", Label: "Keyboard layout", Page: "keyboard", Help: "One layout plus US; Alt+Shift switches for layouts that aren't Latin", Keywords: []string{"layout", "language"}, Command: []string{"qs", "-p", shellPath, "ipc", "call", "menu", "open", "settings.keyboard"}},
		{Name: "open.keys", Label: "Key bindings", Page: "keyboard", Help: "Every key binding, your own included; Super+K shows them too", Keywords: []string{"shortcuts", "bindings"}, Command: []string{"qs", "-p", shellPath, "ipc", "call", "keys", "open"}},
		{Name: "open.themes", Label: "Theme picker", Page: "themes", Help: "Every theme with its preview; Super+Ctrl+T opens it too", Keywords: []string{"theme", "picker"}, Command: []string{"qs", "-p", shellPath, "ipc", "call", "picker", "themes"}},
		{Name: "open.backgrounds", Label: "Wallpaper picker", Page: "wallpaper", Help: "The pictures in use, drawing another, and adding or removing yours", Keywords: []string{"background", "picker"}, Command: []string{"qs", "-p", shellPath, "ipc", "call", "picker", "backgrounds"}},
		{Name: "open.next", Label: "Next wallpaper", Page: "wallpaper", Help: "The next picture along; Super+Ctrl+B does it too", Keywords: []string{"next background"}, Command: []string{"solanum-background", "next"}},
	}
}
