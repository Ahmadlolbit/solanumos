local look = {}

local function setting(path, key)
  local file = io.open(path, "r")
  if not file then
    return nil
  end
  local text = "\n" .. file:read("a"):gsub("^\239\187\191", "")
  file:close()
  local value
  for found in text:gmatch("\n[ \t]*" .. key .. "[ \t]*=([^\n]*)") do
    value = found
  end
  return value and value:match("^%s*(.-)%s*$")
end

local function config_home()
  local config = os.getenv("XDG_CONFIG_HOME")
  if not config or config == "" then
    config = (os.getenv("HOME") or "") .. "/.config"
  end
  return config
end

function look.get(key)
  return setting(config_home() .. "/solanum/look.conf", key) or setting("/etc/solanum/look.conf", key)
end

local function bar_value(path, key)
  local file = io.open(path, "r")
  if not file then
    return nil
  end
  local text = file:read("a")
  file:close()
  return text:match('"' .. key .. '"%s*:%s*"([%w-]+)"')
end

function look.zoom()
  local runtime = os.getenv("XDG_RUNTIME_DIR")
  local file = runtime and runtime ~= "" and io.open(runtime .. "/solanum-zoom", "r")
  if not file then
    return 1
  end
  local factor = tonumber(file:read("l") or "")
  file:close()
  if not factor or factor < 1 or factor > 8 then
    return 1
  end
  return factor
end

function look.bar(key)
  local system = os.getenv("SOLANUM_SHELL_SYSTEM") or "/etc/solanum/shell.json"
  return bar_value(config_home() .. "/solanum/shell.json", key) or bar_value(system, key)
end

return look
