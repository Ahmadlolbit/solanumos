pragma Singleton

import QtQuick
import Quickshell
import Quickshell.Io

Singleton {
  id: arrangement

  readonly property int seconds: Number((Store.rows.find(row => row.name === "screens_wait") || {}).value) || 15
  readonly property var scales: [1, 1.25, 1.5, 1.75, 2, 2.5, 3]
  readonly property var turns: [
    {
      value: "0",
      label: "None"
    },
    {
      value: "90",
      label: "90°"
    },
    {
      value: "180",
      label: "180°"
    },
    {
      value: "270",
      label: "270°"
    }
  ]
  property var screens: []
  property var plan: []
  property bool saved: false
  property int picked: 0
  property int left: 0
  readonly property bool trying: trial.running
  readonly property bool changed: JSON.stringify(plan) !== JSON.stringify(screens)
  readonly property var screen: plan[Math.min(picked, plan.length - 1)] || null
  readonly property var rows: screen ? rowsFor(screen) : []

  function load() {
    reader.running = true;
  }

  function sized(mode) {
    const size = mode.split("@")[0];
    return size === "preferred" ? "" : size;
  }

  function rated(mode) {
    return mode.includes("@") ? mode.split("@")[1] : "";
  }

  function planned(item) {
    return {
      output: item.output,
      description: item.description,
      modes: item.modes || [],
      size: sized(item.mode) || sized((item.modes || [])[0] || "") || item.width + "x" + item.height,
      rate: rated(item.mode) || rated((item.modes || [])[0] || ""),
      x: item.x,
      y: item.y,
      scale: item.scale || 1,
      rotation: item.rotation || 0,
      on: item.on !== false,
      main: !!item.main,
      mirror: item.mirror || ""
    };
  }

  function read(text) {
    let shown = null;
    try {
      shown = JSON.parse(text);
    } catch (error) {
      return;
    }
    screens = (shown.screens || []).map(planned);
    plan = JSON.parse(JSON.stringify(screens));
    saved = !!shown.saved;
    picked = Math.min(picked, Math.max(0, plan.length - 1));
  }

  function sizes(item) {
    const found = [];
    item.modes.forEach(mode => {
      if (!found.includes(sized(mode)))
        found.push(sized(mode));
    });
    return found.sort((one, other) => area(other) - area(one));
  }

  function area(size) {
    const sides = size.split("x").map(Number);
    return sides[0] * sides[1];
  }

  function rates(item) {
    return item.modes.filter(mode => sized(mode) === item.size).map(rated);
  }

  function number(value) {
    return String(Math.round(Number(value) * 100) / 100);
  }

  function rowsFor(item) {
    const others = plan.filter(other => other.output !== item.output).map(other => ({
          value: other.output,
          label: other.output + "  ·  " + other.description
        }));
    const scaled = scales.concat(scales.includes(item.scale) ? [] : [item.scale]).sort((one, other) => one - other);
    return [
      {
        kind: "section",
        name: "section.screen",
        label: item.output + "  ·  " + item.description
      },
      choice("resolution", "Resolution", item.size, sizes(item).map(size => ({
            value: size,
            label: size.replace("x", " × ")
          }))),
      choice("rate", "Refresh rate", item.rate, rates(item).map(rate => ({
            value: rate,
            label: number(rate) + " Hz"
          }))),
      choice("scale", "Scale", String(item.scale), scaled.map(scale => ({
            value: String(scale),
            label: Math.round(scale * 100) + " %"
          }))),
      choice("rotation", "Rotation", String(item.rotation), turns),
      toggle("on", "Use this screen", item.on, "Off leaves it dark while it's plugged in"),
      toggle("main", "Use as main screen", item.main, "Workspace 1 and the cursor start on it"),
      choice("mirror", "Mirror", item.mirror, [
        {
          value: "",
          label: "Off, its own picture"
        }
      ].concat(others))
    ];
  }

  function choice(name, label, value, choices) {
    return {
      kind: "choice",
      name: "screen." + name,
      label: label,
      page: "display",
      whose: "yours",
      value: value,
      choices: choices
    };
  }

  function toggle(name, label, value, help) {
    return {
      kind: "toggle",
      name: "screen." + name,
      label: label,
      page: "display",
      whose: "yours",
      value: value ? "yes" : "no",
      help: help
    };
  }

  function change(name, value) {
    const next = JSON.parse(JSON.stringify(plan));
    const item = next[picked];
    if (!item)
      return;
    const field = name.replace("screen.", "");
    if (field === "resolution") {
      item.size = value;
      item.rate = rates(item).includes(item.rate) ? item.rate : rates(item)[0] || "";
    } else if (field === "main") {
      next.forEach(other => other.main = false);
      item.main = value === "yes";
    } else if (field === "on") {
      item.on = value === "yes";
    } else {
      item[field] = ["scale", "rotation"].includes(field) ? Number(value) : value;
    }
    plan = next;
    pack(picked);
  }

  function pack(moved) {
    plan.forEach((item, at) => {
      if (at === moved || !shown(item))
        return;
      const spot = snapped(at, item.x, item.y, 0);
      const next = JSON.parse(JSON.stringify(plan));
      next[at].x = spot.x;
      next[at].y = spot.y;
      plan = next;
    });
  }

  function box(item) {
    const size = item.size.split("x").map(Number);
    const turned = item.rotation === 90 || item.rotation === 270;
    return {
      x: item.x,
      y: item.y,
      w: Math.round((turned ? size[1] : size[0]) / item.scale),
      h: Math.round((turned ? size[0] : size[1]) / item.scale)
    };
  }

  function shown(item) {
    return item.on && !item.mirror;
  }

  function overlaps(one, other) {
    return one.x < other.x + other.w && other.x < one.x + one.w && one.y < other.y + other.h && other.y < one.y + one.h;
  }

  function lined(value, length, start, end, reach) {
    if (Math.abs(value - start) <= reach)
      return start;
    if (Math.abs(value + length - end) <= reach)
      return end - length;
    return value;
  }

  function beside(mine, other, x, y, reach) {
    const slideX = lined(Math.max(other.x - mine.w, Math.min(other.x + other.w, x)), mine.w, other.x, other.x + other.w, reach);
    const slideY = lined(Math.max(other.y - mine.h, Math.min(other.y + other.h, y)), mine.h, other.y, other.y + other.h, reach);
    return [
      {
        x: other.x + other.w,
        y: slideY
      },
      {
        x: other.x - mine.w,
        y: slideY
      },
      {
        x: slideX,
        y: other.y + other.h
      },
      {
        x: slideX,
        y: other.y - mine.h
      }
    ];
  }

  function snapped(at, x, y, reach) {
    const mine = box(plan[at]);
    const others = plan.filter((item, index) => index !== at && shown(item)).map(box);
    if (!others.length || !shown(plan[at]))
      return {
        x: Math.round(x),
        y: Math.round(y)
      };
    const free = [].concat(...others.map(other => beside(mine, other, x, y, reach))).filter(spot => !others.some(other => overlaps({
            x: spot.x,
            y: spot.y,
            w: mine.w,
            h: mine.h
          }, other)));
    const best = free.sort((one, other) => Math.hypot(one.x - x, one.y - y) - Math.hypot(other.x - x, other.y - y))[0] || {
      x: plan[at].x,
      y: plan[at].y
    };
    return {
      x: Math.round(best.x),
      y: Math.round(best.y)
    };
  }

  function hop(at, dx, dy) {
    const others = plan.filter((item, index) => index !== at && shown(item)).map(box);
    if (!others.length || !plan[at] || !shown(plan[at]))
      return false;
    const mine = box(plan[at]);
    const left = Math.min(...others.map(other => other.x));
    const top = Math.min(...others.map(other => other.y));
    const x = dx < 0 ? left - mine.w : dx > 0 ? Math.max(...others.map(other => other.x + other.w)) : left;
    const y = dy < 0 ? top - mine.h : dy > 0 ? Math.max(...others.map(other => other.y + other.h)) : top;
    drop(at, x, y, 0);
    return true;
  }

  function drop(at, x, y, reach) {
    const spot = snapped(at, x, y, reach);
    const next = JSON.parse(JSON.stringify(plan));
    next[at].x = spot.x;
    next[at].y = spot.y;
    picked = at;
    plan = next;
  }

  function words() {
    const placed = plan.filter(shown).length ? plan.filter(shown) : plan;
    const minX = Math.min(...placed.map(item => item.x));
    const minY = Math.min(...placed.map(item => item.y));
    return [].concat(...plan.map(item => [item.output, "mode=" + item.size + (item.rate ? "@" + item.rate : ""), "position=" + (item.x - minX) + "x" + (item.y - minY), "scale=" + item.scale, "rotation=" + item.rotation, "on=" + (item.on ? "yes" : "no"), "main=" + (item.main ? "yes" : "no"), "mirror=" + (item.mirror || "none")]));
  }

  function apply() {
    if (trying || !plan.length || (!changed && saved))
      return false;
    Store.said = [];
    Store.result = "";
    Store.word = "";
    left = seconds;
    trial.command = ["setsid", "--fork", "--wait", "solanum-look", "screens", "save", "--try", String(seconds)].concat(words());
    trial.running = true;
    return true;
  }

  function answer(keep) {
    if (trying)
      Quickshell.execDetached(["solanum-look", "screens", keep ? "confirm" : "revert"]);
  }

  function undo() {
    plan = JSON.parse(JSON.stringify(screens));
  }

  function ended(code) {
    left = 0;
    if (code === 0 || code === 3) {
      Store.word = code === 0 ? "Kept the new layout" : "Put the old layout back";
      Store.result = "done";
    } else {
      Store.result = "failed";
    }
    load();
  }

  Process {
    id: reader

    command: ["solanum-look", "screens", "now"]
    stdout: StdioCollector {
      onStreamFinished: arrangement.read(text)
    }
  }

  Process {
    id: trial

    stderr: SplitParser {
      onRead: line => Store.heard(line)
    }
    onExited: code => arrangement.ended(code)
  }

  Timer {
    interval: 1000
    repeat: true
    running: trial.running && arrangement.left > 0
    onTriggered: arrangement.left -= 1
  }
}
