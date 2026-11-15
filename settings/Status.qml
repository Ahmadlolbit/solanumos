import QtQuick
import qs.common

Rectangle {
  id: status

  readonly property string last: Store.said.length ? Store.said[Store.said.length - 1] : ""
  readonly property bool shown: !!Store.busy || !!Store.result

  height: shown ? Math.max(Theme.size(44), lines.implicitHeight + 24) : 0
  color: Theme.surface
  clip: true

  Rectangle {
    anchors.left: parent.left
    anchors.right: parent.right
    anchors.top: parent.top
    height: 1
    color: Theme.overlay
  }

  Column {
    id: lines

    anchors.left: parent.left
    anchors.right: close.left
    anchors.verticalCenter: parent.verticalCenter
    anchors.leftMargin: 24
    anchors.rightMargin: 12
    spacing: 4

    Label {
      width: parent.width
      text: Store.busy ? "\uf110  Working on it" : Store.result === "done" ? "\uf00c  " + (Store.word || "Done") : "\uf071  That didn't change" + (status.last ? ": " + status.last : "")
      color: Store.busy ? Theme.accent : Store.result === "done" ? Theme.good : Theme.critical
      font.bold: true
      wrapMode: Text.Wrap
    }

    Repeater {
      model: Store.result === "failed" ? Store.said.slice(-4, -1) : Store.said.slice(-4)

      Label {
        required property string modelData

        width: lines.width
        text: modelData
        color: Theme.subtext
        font.pixelSize: Theme.size(12)
        elide: Text.ElideRight
      }
    }
  }

  Label {
    id: close

    anchors.right: parent.right
    anchors.verticalCenter: parent.verticalCenter
    anchors.rightMargin: 24
    visible: !Store.busy
    text: "\uf00d"
    color: Theme.subtext

    MouseArea {
      anchors.fill: parent
      cursorShape: Qt.PointingHandCursor
      onClicked: Store.dismiss()
    }
  }

  Timer {
    interval: 4000
    running: Store.result === "done" && !Store.busy
    onTriggered: Store.dismiss()
  }
}
