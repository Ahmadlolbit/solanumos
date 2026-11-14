import QtQuick
import qs.common

Rectangle {
  id: tile

  property string title: ""
  property string hint: ""
  property bool current: false
  default property alias body: holder.data

  implicitHeight: strip.height + holder.childrenRect.height + 36
  radius: Theme.radius
  color: "transparent"
  border.width: current ? 2 : 1
  border.color: current ? Theme.accent : Theme.overlay

  Heading {
    id: strip

    anchors.left: parent.left
    anchors.right: parent.right
    anchors.top: parent.top
    anchors.margins: 16
    text: tile.title
    note: tile.hint
  }

  Item {
    id: holder

    anchors.left: parent.left
    anchors.right: parent.right
    anchors.top: strip.bottom
    anchors.margins: 16
    anchors.topMargin: 12
    height: childrenRect.height
  }
}
