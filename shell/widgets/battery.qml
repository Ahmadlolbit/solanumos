import QtQuick
import QtQuick.Layouts
import qs.common
import qs.common.panels

Label {
  id: root

  property var bar
  readonly property bool shown: panel.present

  text: panel.icon() + " " + Math.round(panel.level * 100) + "%"
  color: panel.tint

  MouseArea {
    anchors.fill: parent
    cursorShape: Qt.PointingHandCursor
    onClicked: popup.toggle()
  }

  BarPopup {
    id: popup

    name: "battery"
    bar: root.bar
    owner: root

    Power {
      id: panel

      Layout.fillWidth: true
      active: popup.visible
    }
  }
}
