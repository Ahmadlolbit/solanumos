local look = require("solanum.hypr.look")

hl.on("hyprland.start", function()
  hl.exec_cmd("solanum-session up")
  hl.exec_cmd("dbus-update-activation-environment --systemd --all")
  hl.exec_cmd("xdg-user-dirs-update")
  hl.exec_cmd("/usr/libexec/hyprpolkitagent")
  hl.exec_cmd("udiskie --automount --no-tray")
  hl.exec_cmd("solanum-start hypridle")
  hl.exec_cmd("solanum-start hyprpaper")
  hl.exec_cmd("solanum-start mako")
  hl.exec_cmd("solanum-start elephant")
  hl.exec_cmd("solanum-start walker")
  hl.exec_cmd("solanum-start bar")
  hl.exec_cmd("solanum changes --notify")
  hl.exec_cmd("solanum-crashes login")
  hl.exec_cmd("solanum-look presentation off")
  hl.exec_cmd("solanum-theme-set --editors")
  if look.get("reader") == "yes" then
    hl.exec_cmd("solanum-start orca")
  end
  if look.get("chime") == "yes" then
    hl.exec_cmd("solanum-look chime play")
  end
end)
