local screens = require("solanum.hypr.screens")

local turns = { ["0"] = 0, ["90"] = 1, ["180"] = 2, ["270"] = 3 }

local function given(value, fallback)
  if value == nil or value == "" then
    return fallback
  end
  return value
end

local function rule(screen, outputs)
  local output = outputs[screen.screen]
  if screen.on == "no" then
    return { output = output, disabled = true }
  end
  return {
    output = output,
    mode = given(screen.mode, "preferred"),
    position = given(screen.position, "auto"),
    scale = tonumber(screen.scale) or "auto",
    transform = turns[given(screen.rotation, "0")] or 0,
    mirror = outputs[given(screen.mirror, "")],
  }
end

local function main(output)
  hl.workspace_rule({ workspace = "1", monitor = output, default = true })
  hl.config({ cursor = { default_monitor = output } })
end

local function plugged_now()
  local ok, attached = pcall(screens.attached)
  return ok and screens.signature(attached) or ""
end

local function apply()
  local profile, outputs = screens.active()
  for _, screen in ipairs(profile and profile.screens or {}) do
    hl.monitor(rule(screen, outputs))
    if screen.main == "yes" and screen.on ~= "no" then
      main(outputs[screen.screen])
    end
  end
end

local plugged = plugged_now()
local ok, problem = pcall(apply)
if not ok then
  print("solanum: the screens' layout was left out: " .. tostring(problem))
end

local function replugged()
  if plugged_now() ~= plugged then
    hl.exec_cmd("hyprctl reload")
  end
end

hl.on("monitor.added", replugged)
hl.on("monitor.removed", replugged)
