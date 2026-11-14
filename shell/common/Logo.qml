pragma Singleton

import QtQuick
import Quickshell
import Quickshell.Io

Singleton {
  id: root

  readonly property int size: 32
  readonly property int margin: 6
  readonly property int span: size + 2 * margin
  readonly property int cells: span * 4
  readonly property var bayer: [0, 8, 2, 10, 12, 4, 14, 6, 3, 11, 1, 9, 15, 7, 13, 5]
  property var frames: []
  property int mark: 0
  readonly property bool ready: frames.length > 0

  function lit(rows, x, y) {
    return x >= 0 && y >= 0 && x < size && y < size && ((rows[y] >>> (size - 1 - x)) & 1) === 1;
  }

  function blurred(glow, across) {
    const pass = [];
    for (let at = 0; at < glow.length; at++) {
      const x = at % span, y = Math.floor(at / span);
      let sum = 0;
      for (let k = -2; k <= 2; k++) {
        const xx = across ? x + k : x, yy = across ? y : y + k;
        if (xx >= 0 && yy >= 0 && xx < span && yy < span)
          sum += glow[yy * span + xx];
      }
      pass.push(sum / 5);
    }
    return pass;
  }

  function haloAt(glow, gx, gy, sx, sy) {
    const value = (x, y) => x < span && y < span ? glow[y * span + x] : 0;
    const fx = (sx + 0.5) / 4, fy = (sy + 0.5) / 4;
    const g = (value(gx, gy) * (1 - fx) + value(gx + 1, gy) * fx) * (1 - fy) + (value(gx, gy + 1) * (1 - fx) + value(gx + 1, gy + 1) * fx) * fy;
    const px = gx * 4 + sx, py = gy * 4 + sy;
    return g * 1.35 * 16 > bayer[((py & 3) * 4) + (px & 3)] + 0.5;
  }

  function cellAt(rows, glow, x, y) {
    const gx = Math.floor(x / 4), gy = Math.floor(y / 4);
    if (lit(rows, gx - margin, gy - margin))
      return 1;
    return haloAt(glow, gx, gy, x % 4, y % 4) ? 2 : 0;
  }

  function runsOf(rows) {
    let glow = [];
    for (let at = 0; at < span * span; at++)
      glow.push(lit(rows, at % span - margin, Math.floor(at / span) - margin) ? 1 : 0);
    for (let pass = 0; pass < 2; pass++)
      glow = blurred(blurred(glow, true), false);
    const found = [[], []];
    for (let y = 0; y < cells; y++) {
      let start = 0, kind = 0;
      for (let x = 0; x <= cells; x++) {
        const here = x < cells ? cellAt(rows, glow, x, y) : 0;
        if (here === kind)
          continue;
        if (kind)
          found[kind - 1].push(start, y, x - start);
        start = x;
        kind = here;
      }
    }
    return found;
  }

  function paints(index) {
    const frame = frames[index];
    if (!frame.runs)
      frame.runs = runsOf(frame.rows);
    return frame.runs;
  }

  function load(text) {
    const loaded = [];
    for (const line of text.split("\n")) {
      const fields = line.trim().split(/\s+/);
      const ticks = parseInt(fields[0]);
      if (fields.length !== size + 2 || !(ticks > 0))
        continue;
      if (fields[1] === "logo6")
        root.mark = loaded.length;
      loaded.push({
        ticks: ticks,
        rows: fields.slice(2).map(word => parseInt(word, 16)),
        runs: null
      });
    }
    frames = loaded;
  }

  Process {
    running: true
    command: ["solanum-background", "--frames"]
    stdout: StdioCollector {
      onStreamFinished: root.load(text)
    }
  }
}
