local look = require("solanum.hypr.look")

local filters = { greyscale = true, deuteranopia = true, protanopia = true, tritanopia = true, invert = true }

hl.config({
  cursor = {
    zoom_factor = look.zoom(),
    zoom_rigid = look.get("zoom_follow") == "centre",
  },
})

local filter = look.get("filter")
if filters[filter] then
  hl.config({ decoration = { screen_shader = "/usr/share/solanum/hypr/shaders/" .. filter .. ".frag" } })
end

local animations = look.get("animations")
if animations then
  hl.config({ animations = { enabled = animations ~= "no" } })
end
