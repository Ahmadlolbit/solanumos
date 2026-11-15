import QtQuick
import Quickshell
import Quickshell.Hyprland
import Quickshell.Services.UPower
import Quickshell.Wayland
import qs.common

PanelWindow {
  id: window

  required property var modelData
  property string kind: ""
  readonly property real ratio: devicePixelRatio || 1
  readonly property int wide: Math.round(width * ratio)
  readonly property int high: Math.round(height * ratio)
  readonly property HyprlandMonitor monitor: Hyprland.monitorFor(modelData)
  readonly property var workspace: monitor ? monitor.activeWorkspace : null
  readonly property bool covered: !!workspace && (workspace.hasFullscreen || workspace.toplevels.values.length > 0)
  readonly property bool resting: covered || UPower.onBattery || Screens.dark

  screen: modelData
  color: Theme.base
  exclusionMode: ExclusionMode.Ignore
  mask: Region {}
  WlrLayershell.layer: WlrLayer.Bottom
  WlrLayershell.namespace: kind

  anchors {
    top: true
    bottom: true
    left: true
    right: true
  }
}
