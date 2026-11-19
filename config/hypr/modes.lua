local look = require("solanum.hypr.look")
local screens = require("solanum.hypr.screens")

if look.get("zen") == "yes" then
  hl.config({
    general = { gaps_in = 0, gaps_out = 0 },
    decoration = { rounding = 0 },
  })
end

if look.get("presentation") == "yes" and not screens.lid_closed() then
  local panels, others = screens.showing()
  for _, output in ipairs(others) do
    if panels[1] then
      hl.monitor({ output = output, mirror = panels[1] })
    end
  end
end
