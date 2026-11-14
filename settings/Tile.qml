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

  Item {
    id: strip

    anchors.left: parent.left
    anchors.right: parent.right
    anchors.top: parent.top
    anchors.margins: 16
    height: Theme.size(20)

    Label {
      anchors.left: parent.left
      anchors.verticalCenter: parent.verticalCenter
      text: tile.title.toUpperCase()
      color: Theme.subtext
      font.pixelSize: Theme.size(11)
      font.bold: true
    }

    Label {
      anchors.right: parent.right
      anchors.verticalCenter: parent.verticalCenter
      width: Math.min(implicitWidth, parent.width * 0.7)
      text: tile.hint
      color: Theme.subtext
      font.pixelSize: Theme.size(11)
      elide: Text.ElideLeft
      horizontalAlignment: Text.AlignRight
    }
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
