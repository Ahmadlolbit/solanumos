import QtQuick
import qs.common

Canvas {
  id: mark

  property string kind: "skip"
  property color ink: Theme.text
  property int phase: 0
  readonly property var patterns: ({
      ok: () => true,
      fail: (x, y) => (x + y) % 3 === 0,
      warn: (x, y, cell) => (Math.floor(x / cell) + Math.floor(y / cell)) % 2 === 0,
      busy: (x, y, cell, shift) => (Math.floor((x + shift) / cell) + Math.floor(y / cell)) % 2 === 0
    })

  implicitWidth: Theme.size(12)
  implicitHeight: implicitWidth
  antialiasing: false
  smooth: false
  onKindChanged: requestPaint()
  onInkChanged: requestPaint()
  onPhaseChanged: requestPaint()
  onWidthChanged: requestPaint()
  onPaint: {
    const context = getContext("2d");
    const size = Math.round(width), cell = Math.max(1, Math.round(size / 6));
    const pattern = patterns[kind] || (() => false);
    context.reset();
    context.fillStyle = ink;
    for (let y = 0; y < size; y++)
      for (let x = 0; x < size; x++)
        if (x === 0 || y === 0 || x === size - 1 || y === size - 1 || pattern(x, y, cell, phase))
          context.fillRect(x, y, 1, 1);
  }

  Timer {
    interval: 120
    running: mark.visible && mark.kind === "busy"
    repeat: true
    onTriggered: mark.phase = (mark.phase + 1) % 120
  }
}
