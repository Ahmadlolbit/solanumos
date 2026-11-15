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
	Group    string   `json:"group"`
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
	{"themes", "Appearance and themes", "\uf53f", "Look", []string{"theme", "colours", "colors", "accent", "dark", "light", "fonts", "style", "your own theme"}},
	{"wallpaper", "Wallpaper and desktop", "\uf03e", "Look", []string{"wallpaper", "background", "pictures", "slideshow", "logo", "life"}},
	{"bar", "Bar, dock and menu", "\uf0c9", "Look", []string{"bar", "widgets", "clock", "weather", "zen"}},
	{"windows", "Windows and workspaces", "\uf2d2", "Look", []string{"windows", "animations", "workspaces"}},
	{"display", "Display", "\uf108", "Devices", []string{"screen", "scale", "brightness", "night light", "corners"}},
	{"sound", "Sound", "\uf028", "Devices", []string{"audio", "volume", "speakers", "microphone", "chime"}},
	{"network", "Network", "\uf1eb", "Devices", []string{"wi-fi", "wifi", "internet", "networkmanager"}},
	{"bluetooth", "Bluetooth", "\uf293", "Devices", []string{"devices", "pair", "headphones"}},
	{"power", "Power and battery", "\uf240", "Devices", []string{"battery", "power profile", "suspend", "hibernate", "charge"}},
	{"keyboard", "Keyboard", "\uf11c", "Devices", []string{"layout", "keys", "shortcuts", "bindings"}},
	{"mouse", "Mouse and touchpad", "\uf245", "Devices", []string{"touchpad", "trackpoint", "scrolling", "tap"}},
	{"hardware", "Hardware and drivers", "\uf2db", "Devices", []string{"kernel", "graphics", "nvidia", "webcam", "drivers"}},
	{"lock", "Lock screen and idle", "\uf023", "You", []string{"lock", "idle", "screensaver", "login screen"}},
	{"notifications", "Notifications", "\uf0f3", "You", []string{"do not disturb", "quiet hours", "tips", "warnings"}},
	{"apps", "Apps and default apps", "\uf009", "You", []string{"apps", "applications", "flatpak", "extras", "profile"}},
	{"accounts", "Accounts and family", "\uf007", "You", []string{"fingerprint", "account", "user"}},
	{"privacy", "Privacy and security", "\uf132", "You", []string{"firewall", "security"}},
	{"sharing", "Sharing and remote access", "\uf1e0", "You", []string{"ssh", "remote"}},
	{"time", "Date and time", "\uf017", "You", []string{"clock", "24-hour", "12-hour", "location", "sun"}},
	{"accessibility", "Accessibility", "\uf29a", "You", []string{"screen reader", "orca", "text size", "larger text"}},
	{"updates", "Updates, snapshots and backups", "\uf021", "System", []string{"updates", "snapshots", "restore", "mirror"}},
	{"storage", "Storage and memory", "\uf1c0", "System", []string{"swap", "zram", "memory", "disk"}},
	{"startup", "Startup, login and boot", "\uf011", "System", []string{"boot", "splash", "grub", "console"}},
	{"developer", "Developer", "\uf121", "System", []string{"docker", "podman", "mise", "prompt", "containers"}},
	{"about", "About", "\uf05a", "System", []string{"version", "system", "hardware", "credits", "licences", "debian"}},
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
