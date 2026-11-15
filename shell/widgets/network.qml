import QtQuick
import QtQuick.Layouts
import Quickshell.Networking
import qs.common
import qs.common.panels

Label {
  id: root

  property var bar
  readonly property var device: {
    const devices = Networking.devices.values;
    return devices.find(item => item.connected && item.type === DeviceType.Wired) || devices.find(item => item.connected && item.type === DeviceType.Wifi) || null;
  }
  readonly property var network: device && device.type === DeviceType.Wifi ? device.networks.values.find(item => item.connected) || null : null

  function icon() {
    if (!device)
      return "\u{f092e}";
    if (device.type === DeviceType.Wired)
      return "\u{f0200}";
    return panel.signalIcon(network ? panel.strength(network) : 0);
  }

  text: network && !(bar && bar.vertical) ? icon() + " " + network.name : icon()
  color: device ? Theme.text : Theme.subtext

  MouseArea {
    anchors.fill: parent
    cursorShape: Qt.PointingHandCursor
    onClicked: popup.toggle()
  }

  BarPopup {
    id: popup

    name: "wifi"
    bar: root.bar
    owner: root

    Wifi {
      id: panel

      Layout.fillWidth: true
      active: popup.visible
      onDone: Popups.close(popup)
    }
  }
}
