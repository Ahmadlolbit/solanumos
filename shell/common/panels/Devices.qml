import QtQuick
import QtQuick.Layouts
import Quickshell
import Quickshell.Bluetooth
import qs.common

ColumnLayout {
  id: root

  property bool active: false
  readonly property var adapter: Bluetooth.defaultAdapter
  readonly property var devices: adapter ? adapter.devices.values : []
  readonly property bool linked: devices.some(device => device.connected)

  signal done

  function deviceIcon(device) {
    const kind = device.icon || "";
    if (kind.startsWith("audio"))
      return "\u{f02cb}";
    if (kind === "input-keyboard")
      return "\u{f030c}";
    if (kind === "input-mouse")
      return "\u{f037d}";
    if (kind === "input-gaming")
      return "\u{f0297}";
    if (kind === "phone")
      return "\u{f03f2}";
    return "\u{f00af}";
  }

  function detail(device) {
    if (device.pairing)
      return "Pairing...";
    if (device.state === BluetoothDeviceState.Connecting)
      return "Connecting...";
    if (device.connected)
      return device.batteryAvailable ? Math.round(device.battery * 100) + "%" : "Connected";
    return device.paired ? "Paired" : "";
  }

  function choose(device) {
    if (device.connected)
      device.disconnect();
    else if (device.paired)
      device.connect();
    else
      device.pair();
  }

  spacing: 10
  onActiveChanged: {
    if (!active && adapter && adapter.discovering)
      adapter.discovering = false;
  }

  RowLayout {
    Layout.fillWidth: true
    Layout.preferredWidth: Theme.size(300)
    spacing: 10

    Label {
      Layout.fillWidth: true
      text: "Bluetooth"
      font.bold: true
    }

    Toggle {
      visible: root.adapter !== null
      worded: true
      checked: root.adapter ? root.adapter.enabled : false
      onToggled: root.adapter.enabled = !root.adapter.enabled
    }
  }

  Label {
    visible: !list.visible
    text: !root.adapter ? "No Bluetooth adapter" : !root.adapter.enabled ? "Bluetooth is off" : root.adapter.discovering ? "Looking for devices..." : "No devices yet; Scan finds new ones"
    color: Theme.subtext
  }

  Rows {
    id: list

    Layout.fillWidth: true
    Layout.preferredHeight: Math.min(contentHeight, 300)
    spacing: 2
    model: ScriptModel {
      values: root.adapter && root.adapter.enabled ? root.devices.filter(device => device.paired || root.adapter.discovering).sort((a, b) => b.connected - a.connected || b.paired - a.paired || a.name.localeCompare(b.name)) : []
    }

    delegate: Choice {
      id: choice

      required property var modelData

      width: list.width
      activeFocusOnTab: false
      icon: root.deviceIcon(modelData)
      label: modelData.name || modelData.address
      detail: root.detail(modelData)
      selected: modelData.connected
      action: modelData.paired ? "Forget" : ""
      onClicked: root.choose(modelData)
      onActionClicked: modelData.forget()

      Connections {
        target: choice.modelData

        function onPairedChanged() {
          if (!choice.modelData.paired)
            return;
          choice.modelData.trusted = true;
          choice.modelData.connect();
        }
      }
    }
  }

  Choice {
    Layout.fillWidth: true
    visible: root.adapter !== null && root.adapter.enabled
    icon: "\u{f002}"
    label: root.adapter && root.adapter.discovering ? "Stop scanning" : "Scan"
    onClicked: root.adapter.discovering = !root.adapter.discovering
  }

  Choice {
    Layout.fillWidth: true
    icon: "\u{f013}"
    label: "More settings"
    onClicked: {
      root.done();
      Quickshell.execDetached(["blueman-manager"]);
    }
  }
}
