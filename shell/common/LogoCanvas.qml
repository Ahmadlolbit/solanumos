import QtQuick

Canvas {
  id: canvas

  property real cell: 4
  property bool resting: false
  property int shown: Logo.mark
  property var pictures: []
  readonly property var paints: [Theme.base, Theme.text, Theme.accent]
  readonly property bool holding: Logo.ready && Logo.frames[shown].ticks > 1
  readonly property bool playing: Logo.ready && visible && (!resting || !holding)

  function picture(context) {
    if (!pictures[shown]) {
      const made = context.createImageData(Logo.cells, Logo.cells);
      const grid = Logo.grid(shown), data = made.data;
      const channels = paints.map(paint => [paint.r, paint.g, paint.b].map(part => Math.round(part * 255)));
      for (let at = 0; at < grid.length; at++) {
        const paint = channels[grid[at]];
        data[at * 4] = paint[0];
        data[at * 4 + 1] = paint[1];
        data[at * 4 + 2] = paint[2];
        data[at * 4 + 3] = 255;
      }
      pictures[shown] = made;
    }
    return pictures[shown];
  }

  width: Logo.cells
  height: Logo.cells
  scale: cell
  transformOrigin: Item.TopLeft
  smooth: false
  antialiasing: false
  onShownChanged: requestPaint()
  onPaintsChanged: {
    pictures = [];
    requestPaint();
  }
  onPaint: {
    const context = getContext("2d");
    if (Logo.ready)
      context.drawImage(picture(context), 0, 0);
  }

  Connections {
    target: Logo

    function onReadyChanged() {
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
