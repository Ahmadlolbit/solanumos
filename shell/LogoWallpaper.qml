import QtQuick
import Quickshell
import qs.common

Variants {
  model: Quickshell.screens

  LiveWindow {
    id: window

    readonly property int scale: 4 * Math.max(1, Math.round(4 * high / 1600))

    kind: "solanum-logo"
    visible: Logo.ready

    Rectangle {
      anchors.fill: parent
      color: Theme.base

      LogoCanvas {
        x: (Math.trunc((window.wide - Logo.size * window.scale) / 2) - Logo.margin * window.scale) / window.ratio
        y: (Math.trunc((window.high - Logo.size * window.scale) / 2) - Logo.margin * window.scale) / window.ratio
        cell: window.scale / 4 / window.ratio
        resting: window.resting
      }
    }
  }
}
