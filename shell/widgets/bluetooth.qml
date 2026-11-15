import QtQuick
import QtQuick.Layouts
import qs.common
import qs.common.panels

Label {
  id: root

  property var bar
  readonly property bool shown: panel.adapter !== null

  function icon() {
    if (!panel.adapter || !panel.adapter.enabled)
      return "\u{f00b2}";
    return panel.linked ? "\u{f00b1}" : "\u{f00af}";
  }

  text: icon()
  color: panel.adapter && panel.adapter.enabled ? Theme.text : Theme.subtext

  MouseArea {
    anchors.fill: parent
    cursorShape: Qt.PointingHandCursor
    onClicked: popup.toggle()
  }

  BarPopup {
    id: popup

    name: "bluetooth"
    bar: root.bar
    owner: root

    Devices {
      id: panel

      Layout.fillWidth: true
      active: popup.visible
      onDone: Popups.close(popup)
    }
  }
}
