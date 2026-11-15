import QtQuick
import Quickshell
import qs.common

Variants {
  model: Quickshell.screens

  LiveWindow {
    id: window

    readonly property int pitch: Math.max(1, Math.round(high / 90))
    readonly property int cols: Math.max(1, Math.floor(wide / pitch))
    readonly property int rows: Math.max(1, Math.floor(high / pitch))

    kind: "solanum-life"
    color: player.ready ? Theme.base : "transparent"

    Rectangle {
      anchors.fill: parent
      color: Theme.base
      visible: player.ready

      LifePlayer {
        id: player

        x: Math.trunc((window.wide - window.cols * window.pitch) / 2) / window.ratio
        y: Math.trunc((window.high - window.rows * window.pitch) / 2) / window.ratio
        name: window.modelData.name
        cols: window.cols
        rows: window.rows
        scale: window.pitch / window.ratio
        resting: window.resting
      }
    }
  }
}
