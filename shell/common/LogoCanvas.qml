import QtQuick

Canvas {
  id: canvas

  property real cell: 4
  property bool resting: false
  property int shown: Logo.mark
  readonly property bool holding: Logo.ready && Logo.frames[shown].ticks > 1
  readonly property bool playing: Logo.ready && visible && (!resting || !holding)

  width: Logo.cells * cell
  height: Logo.cells * cell
  antialiasing: false
  onShownChanged: requestPaint()
  onCellChanged: requestPaint()
  onPaint: {
    const context = getContext("2d");
    context.fillStyle = Theme.base;
    context.fillRect(0, 0, width, height);
    if (!Logo.ready)
      return;
    const found = Logo.paints(shown);
    [Theme.text, Theme.accent].forEach((paint, kind) => {
      context.fillStyle = paint;
      const runs = found[kind];
      for (let at = 0; at < runs.length; at += 3)
        context.fillRect(runs[at] * cell, runs[at + 1] * cell, runs[at + 2] * cell, cell);
    });
  }

  Connections {
    target: Theme

    function onBaseChanged() {
      canvas.requestPaint();
    }

    function onTextChanged() {
      canvas.requestPaint();
    }

    function onAccentChanged() {
      canvas.requestPaint();
    }
  }

  Timer {
    interval: Logo.ready ? Math.max(1, Math.round(Logo.frames[canvas.shown].ticks * 1000 / 60)) : 1000
    running: canvas.playing
    repeat: true
    onTriggered: canvas.shown = (canvas.shown + 1) % Logo.frames.length
  }
}
