import QtQuick

Rectangle {
  id: root

  property string key: ""

  implicitWidth: Math.max(implicitHeight, cap.implicitWidth + 12)
  implicitHeight: cap.implicitHeight + 4
  radius: Math.min(Theme.radius, 4)
  color: "transparent"
  border.width: 1
  border.color: Theme.subtext

  Label {
    id: cap

    anchors.centerIn: parent
    text: root.key
  }
}
