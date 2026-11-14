local state = (os.getenv("HOME") or "") .. "/.local/state"
local roots = "/etc/?.lua;/usr/share/?.lua;" .. state .. "/?.lua;"
if not package.path:find(roots, 1, true) then
  package.path = roots .. package.path
end

require("solanum.hypr.envs")
require("solanum.hypr.looknfeel")
require("solanum.hypr.input")
require("solanum.hypr.windows")
require("solanum.hypr.autostart")
require("solanum.hypr.bindings")

if package.searchpath("solanum.current.theme.hyprland", package.path) then
  require("solanum.current.theme.hyprland")
end

if package.searchpath("solanum.hypr.options", package.path) then
  require("solanum.hypr.options")
end

require("solanum.hypr.lid")
require("solanum.hypr.modes")
require("solanum.hypr.trackpoint")

if package.searchpath("solanum.current.theme.glass", package.path) then
  require("solanum.current.theme.glass")
end
