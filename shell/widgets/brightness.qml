import QtQuick
import QtQuick.Layouts
import qs.common
import qs.common.panels

Label {
  id: root

  property var bar
  readonly property bool shown: Brightness.available

  text: "\u{f00df} " + Math.round(Brightness.level * 100) + "%"

  MouseArea {
    anchors.fill: parent
    cursorShape: Qt.PointingHandCursor
    onClicked: popup.toggle()
    onWheel: wheel => {
      const steps = notches.steps(wheel);
      if (steps !== 0)
        Brightness.adjust(5 * Math.abs(steps) + (steps > 0 ? "%+" : "%-"));
    }

    Notches {
      id: notches
    }
  }

  BarPopup {
    id: popup

    name: "brightness"
    bar: root.bar
    owner: root

    Light {
      Layout.fillWidth: true
    }
  }
}
