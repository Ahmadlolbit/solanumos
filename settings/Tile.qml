import QtQuick
import qs.common

Rectangle {
  id: tile

  property string title: ""
  property string hint: ""
  property bool current: false
  default property alias body: holder.data
  readonly property bool beside: hint === "" || named.advanceWidth + noted.advanceWidth + 16 <= width - 32

  implicitHeight: strip.height + (under.visible ? under.height + 4 : 0) + holder.childrenRect.height + 36
  radius: Theme.radius
  color: "transparent"
  border.width: current ? 2 : 1
  border.color: current ? Theme.accent : Theme.overlay

  TextMetrics {
    id: named

    font.family: Theme.display
    font.pixelSize: Theme.size(11)
    font.bold: Theme.displayBold
    font.capitalization: Font.AllUppercase
    font.letterSpacing: Theme.size(11) * 0.08
    text: tile.title
  }

  TextMetrics {
    id: noted

    font.family: Theme.font
    font.pixelSize: Theme.size(11)
    text: tile.hint
  }

  Heading {
    id: strip

    anchors.left: parent.left
    anchors.right: parent.right
    anchors.top: parent.top
    anchors.margins: 16
    text: tile.title
    note: tile.beside ? tile.hint : ""
  }

  Label {
    id: under

    anchors.left: parent.left
    anchors.right: parent.right
    anchors.top: strip.bottom
    anchors.leftMargin: 16
    anchors.rightMargin: 16
    anchors.topMargin: 4
    visible: !tile.beside
    text: tile.hint
    color: Theme.subtext
    font.pixelSize: Theme.size(11)
    wrapMode: Text.Wrap
  }

  Item {
    id: holder

    anchors.left: parent.left
    anchors.right: parent.right
    anchors.top: under.visible ? under.bottom : strip.bottom
    anchors.margins: 16
    anchors.topMargin: 12
    height: childrenRect.height
  }
}
