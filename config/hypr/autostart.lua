local look = require("solanum.hypr.look")

local function start(command)
  hl.exec_cmd("env -u HL_INITIAL_WORKSPACE_TOKEN " .. command)
end

hl.on("hyprland.start", function()
  start("solanum-session up")
  start("solanum-look screens settle")
  start("dbus-update-activation-environment --systemd --all")
  start("xdg-user-dirs-update")
  start("/usr/libexec/hyprpolkitagent")
  start("udiskie --automount --no-tray")
  start("solanum-start hypridle")
  start("solanum-start hyprpaper")
  start("solanum-start mako")
  start("solanum-start elephant")
  start("solanum-start walker")
  start("solanum-start bar")
  start("solanum changes --notify")
  start("solanum-crashes login")
  start("solanum-look presentation off")
  start("solanum-theme-set --editors")
  if look.get("reader") == "yes" then
    start("solanum-start orca")
  end
  if look.get("chime") == "yes" then
    start("solanum-look chime play")
  end
end)
