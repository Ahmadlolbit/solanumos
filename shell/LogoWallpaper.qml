import QtQuick
import Quickshell
import Quickshell.Hyprland
import Quickshell.Services.UPower
import Quickshell.Wayland
import qs.common

Variants {
  model: Quickshell.screens

  PanelWindow {
    id: window

    required property var modelData
    readonly property real ratio: modelData.devicePixelRatio || 1
    readonly property int wide: Math.round(modelData.width * ratio)
    readonly property int high: Math.round(modelData.height * ratio)
    readonly property int scale: 4 * Math.max(1, Math.round(4 * high / 1600))
    readonly property HyprlandMonitor monitor: Hyprland.monitorFor(modelData)
    readonly property var workspace: monitor ? monitor.activeWorkspace : null
    readonly property bool covered: !!workspace && (workspace.hasFullscreen || workspace.toplevels.values.length > 0)

    screen: modelData
    color: Theme.base
    exclusionMode: ExclusionMode.Ignore
    mask: Region {}
    WlrLayershell.layer: WlrLayer.Bottom
    WlrLayershell.namespace: "solanum-logo"

    anchors {
      top: true
      bottom: true
      left: true
      right: true
    }

    LogoCanvas {
      x: (Math.trunc((window.wide - Logo.size * window.scale) / 2) - Logo.margin * window.scale) / window.ratio
      y: (Math.trunc((window.high - Logo.size * window.scale) / 2) - Logo.margin * window.scale) / window.ratio
      cell: window.scale / 4 / window.ratio
      resting: window.covered || UPower.onBattery || Screens.dark
    }
  }
}
