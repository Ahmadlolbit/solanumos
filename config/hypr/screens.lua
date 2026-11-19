local screens = {}

local root = os.getenv("SOLANUM_SYS_ROOT") or ""

function screens.first_line(path)
  local file = io.open(root .. path, "r")
  if not file then
    return nil
  end
  local line = file:read("l")
  file:close()
  return line
end

local function internal(output)
  return output:match("^eDP") or output:match("^LVDS") or output:match("^DSI")
end

local function plugged()
  local found = {}
  local listing = io.popen("ls " .. root .. "/sys/class/drm 2>/dev/null")
  if not listing then
    return found
  end
  for entry in listing:lines() do
    local output = entry:match("^card%d+%-(.+)$")
    if output and screens.first_line("/sys/class/drm/" .. entry .. "/status") == "connected" then
      table.insert(found, { entry = entry, output = output })
    end
  end
  listing:close()
  return found
end

function screens.connected()
  local panels, others = {}, {}
  for _, screen in ipairs(plugged()) do
    table.insert(internal(screen.output) and panels or others, screen.output)
  end
  return panels, others
end

local function trimmed(text)
  return text:match("^%s*(.-)%s*$")
end

local function escaped(character)
  return string.format("\\x%02x", character:byte())
end

local function descriptor_text(data, at)
  local text = data:sub(at + 6, at + 18):match("^[^\n]*")
  return trimmed((text:gsub("[^\32-\126]", escaped)))
end

local function descriptors(data)
  local found = {}
  for at = 54, 108, 18 do
    local text = data:byte(at + 1) == 0 and data:byte(at + 2) == 0 and descriptor_text(data, at) or ""
    local tag = data:byte(at + 4)
    if text ~= "" and not found[tag] then
      found[tag] = text
    end
  end
  return found
end

function screens.identity(entry)
  local file = io.open(root .. "/sys/class/drm/" .. entry .. "/edid", "rb")
  local data = file and file:read(128) or ""
  if file then
    file:close()
  end
  if #data < 128 then
    return "", ""
  end
  local found = descriptors(data)
  local number = string.unpack("<I4", data, 13)
  local model = found[0xFC] or string.format("0x%04X", string.unpack("<I2", data, 11))
  local serial = found[0xFF] or (number ~= 0 and string.format("0x%08X", number) or "")
  return model, serial
end

function screens.attached()
  local found = {}
  for _, screen in ipairs(plugged()) do
    local model, serial = screens.identity(screen.entry)
    table.insert(found, { output = screen.output, model = model, serial = serial })
  end
  return found
end

function screens.layout_file()
  local config = os.getenv("XDG_CONFIG_HOME")
  if not config or config == "" then
    config = (os.getenv("HOME") or "") .. "/.config"
  end
  return config .. "/solanum/screens.conf"
end

local function read_line(state, line)
  local key = line:match("^%s*%[(.*)%]%s*$")
  local name, value = line:match("^%s*([%w_]+)%s*=%s*(.-)%s*$")
  if key then
    state.profile = { key = key, screens = {} }
    state.screen = nil
    table.insert(state.found, state.profile)
  elseif name == "screen" and state.profile then
    state.screen = { screen = value }
    table.insert(state.profile.screens, state.screen)
  elseif name and state.screen then
    state.screen[name] = value
  end
end

function screens.profiles(path)
  local state = { found = {} }
  local file = io.open(path, "r")
  if not file then
    return state.found
  end
  for line in file:lines() do
    read_line(state, (line:gsub("^\239\187\191", "")))
  end
  file:close()
  return state.found
end

local function take(screen, attached, used, same_output)
  for at, here in ipairs(attached) do
    local same = here.model == (screen.model or "") and here.serial == (screen.serial or "")
    if same and not used[at] and (not same_output or here.output == screen.screen) then
      used[at] = true
      return here.output
    end
  end
end

local function placed(profile, attached)
  local used, outputs = {}, {}
  for _, same_output in ipairs({ true, false }) do
    for _, screen in ipairs(profile.screens) do
      outputs[screen.screen] = outputs[screen.screen] or take(screen, attached, used, same_output)
    end
  end
  for _, screen in ipairs(profile.screens) do
    if not outputs[screen.screen] then
      return nil
    end
  end
  return outputs
end

function screens.signature(attached)
  local names = {}
  for _, screen in ipairs(attached) do
    table.insert(names, table.concat({ screen.output, screen.model, screen.serial }, " "))
  end
  table.sort(names)
  return table.concat(names, "\n")
end

function screens.saved(attached)
  for _, profile in ipairs(screens.profiles(screens.layout_file())) do
    local outputs = #profile.screens == #attached and placed(profile, attached)
    if outputs then
      return profile, outputs
    end
  end
end

local function within(text, low, high)
  local number = tonumber(text or "")
  return number ~= nil and number >= low and number <= high
end

local function good_mode(mode)
  local wide, high, rate = mode:match("^(%d+)x(%d+)@?([%d.]*)$")
  return mode == "" or mode == "preferred"
    or (within(wide, 1, 99999) and within(high, 1, 99999) and (rate == "" and not mode:find("@", 1, true) or within(rate, 1, 999)))
end

local function good_position(position)
  local x, y = position:match("^(%-?%d+)x(%-?%d+)$")
  return position == "" or (within(x, -999999, 999999) and within(y, -999999, 999999))
end

local checks = {
  scale = function(value) return value == "" or within(value, 0.25, 4) end,
  mode = good_mode,
  position = good_position,
  rotation = function(value) return value == "" or value == "0" or value == "90" or value == "180" or value == "270" end,
  on = function(value) return value == "" or value == "yes" or value == "no" end,
  main = function(value) return value == "" or value == "yes" or value == "no" end,
}

local function bad_setting(screen, outputs)
  for name, good in pairs(checks) do
    if not good(screen[name] or "") then
      return name
    end
  end
  local mirror = screen.mirror or ""
  if mirror ~= "" and (mirror == screen.screen or not outputs[mirror]) then
    return "mirror"
  end
end

local function checked(profile, outputs)
  for _, screen in ipairs(profile.screens) do
    local bad = bad_setting(screen, outputs)
    if bad then
      error("a bad " .. bad .. " for " .. tostring(screen.screen), 0)
    end
  end
  return profile, outputs
end

local active

function screens.active()
  if active == nil then
    local ok, profile, outputs = pcall(function()
      local found, places = screens.saved(screens.attached())
      if found then
        return checked(found, places)
      end
    end)
    if not ok then
      print("solanum: screens.conf left out, the default layout applies: " .. tostring(profile))
    end
    active = ok and profile and { profile = profile, outputs = outputs } or false
  end
  if active then
    return active.profile, active.outputs
  end
end

local function role(output)
  local profile, outputs = screens.active()
  for _, screen in ipairs(profile and profile.screens or {}) do
    if outputs[screen.screen] == output then
      return screen.on ~= "no" and not outputs[screen.mirror or ""]
    end
  end
  return true
end

local function own_picture(list)
  local found = {}
  for _, output in ipairs(list) do
    if role(output) then
      table.insert(found, output)
    end
  end
  return found
end

function screens.showing()
  local panels, others = screens.connected()
  return own_picture(panels), own_picture(others)
end

function screens.lid_closed()
  for _, name in ipairs({ "LID", "LID0", "LID1" }) do
    local state = screens.first_line("/proc/acpi/button/lid/" .. name .. "/state")
    if state then
      return state:find("closed", 1, true) ~= nil
    end
  end
  return false
end

return screens
