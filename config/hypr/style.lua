local look = require("solanum.hypr.look")

local style = look.get("style")
if style ~= "soft" then
  style = "sharp"
end

require("solanum.hypr.style-" .. style)
