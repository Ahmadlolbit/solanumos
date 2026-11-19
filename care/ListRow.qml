import QtQuick
import qs.common

Rectangle {
  id: row

  property string title: ""
  property string detail: ""
  property string word: ""
  property bool current: false
  property bool framed: false

  signal picked

  implicitHeight: Math.max(Theme.size(44), texts.implicitHeight + 16)
  radius: Theme.radius
  color: current ? Theme.surface : area.containsMouse ? Qt.alpha(Theme.surface, 0.5) : "transparent"
  border.width: current && framed ? Theme.focusBorder : 0
  border.color: Theme.accent

  Column {
    id: texts

    anchors.left: parent.left
    anchors.leftMargin: 16
    anchors.right: badge.left
    anchors.rightMargin: 16
    anchors.verticalCenter: parent.verticalCenter
    spacing: 2

    Label {
      width: parent.width
      text: row.title
      elide: Text.ElideRight
      font.features: ({
          tnum: 1
        })
    }

    Label {
      width: parent.width
      visible: text !== ""
      text: row.detail
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
    color: "transparent"
    border.width: 1
    border.color: Theme.overlay

    Label {
      id: shown

      anchors.centerIn: parent
      text: row.word.toUpperCase()
      font.pixelSize: Theme.size(11)
      color: Theme.subtext
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
