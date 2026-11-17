import QtQuick
import qs.common

Rectangle {
  id: row

  property var item: ({})
  property bool current: false
  property bool marked: false
  readonly property bool changeable: ["app", "pkg"].includes(item.kind)
  readonly property bool on: !!item.installed
  readonly property string word: marked ? (on ? "Will remove" : "Will install") : item.kind === "web" ? "" : on ? "Installed" : changeable ? "Not installed" : ""
  readonly property string glyph: marked ? "▣" : on ? "■" : "□"

  signal picked
  signal toggled

  implicitHeight: Math.max(Theme.size(44), texts.implicitHeight + 16)
  radius: Theme.radius
  color: current ? Theme.surface : area.containsMouse ? Qt.alpha(Theme.surface, 0.5) : "transparent"
  border.width: current ? Theme.focusBorder : 0
  border.color: Theme.accent

  Label {
    id: box

    anchors.left: parent.left
    anchors.leftMargin: 16
    anchors.verticalCenter: parent.verticalCenter
    visible: row.changeable
    text: row.glyph
    color: row.present || row.marked ? Theme.text : Theme.subtext

    MouseArea {
      anchors.fill: parent
      anchors.margins: -8
      cursorShape: Qt.PointingHandCursor
      onClicked: row.toggled()
    }
  }

  Column {
    id: texts

    anchors.left: row.changeable ? box.right : parent.left
    anchors.leftMargin: 16
    anchors.right: badge.left
    anchors.rightMargin: 16
    anchors.verticalCenter: parent.verticalCenter
    spacing: 2

    Label {
      width: parent.width
      text: row.item.label || row.item.name || ""
      font.bold: row.item.kind === "add"
      color: row.item.kind === "note" ? Theme.subtext : Theme.text
      elide: Text.ElideRight
    }

    Label {
      width: parent.width
      visible: text !== ""
      text: row.item.detail || ""
      color: Theme.subtext
      font.pixelSize: Theme.size(11)
      elide: Text.ElideRight
    }
  }

  Rectangle {
    id: badge

    anchors.right: parent.right
    anchors.rightMargin: 16
    anchors.verticalCenter: parent.verticalCenter
    visible: row.word !== ""
    width: visible ? shown.implicitWidth + 16 : 0
    height: shown.implicitHeight + 6
    radius: Theme.radius
    color: row.marked ? Theme.text : "transparent"
    border.width: 1
    border.color: row.present || row.marked ? Theme.text : Theme.overlay

    Label {
      id: shown

      anchors.centerIn: parent
      text: row.word.toUpperCase()
      font.pixelSize: Theme.size(11)
      color: row.marked ? Theme.base : row.present ? Theme.text : Theme.subtext
    }
  }

  MouseArea {
    id: area

    anchors.fill: parent
    z: -1
    hoverEnabled: true
    cursorShape: Qt.PointingHandCursor
    onClicked: row.picked()
  }
}
