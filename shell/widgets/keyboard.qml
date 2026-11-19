import QtQuick
import Quickshell
import Quickshell.Io
import qs.common

Label {
  id: root

  property var bar
  property bool touch: false
  readonly property bool shown: touch || Session.keyboardShown

  text: "\u{f030c}"
  color: Session.keyboardShown ? Theme.accent : Theme.text

  Process {
    running: true
    command: ["hyprctl", "devices", "-j"]
    stdout: StdioCollector {
      onStreamFinished: {
        try {
          root.touch = (JSON.parse(text).touch || []).length > 0;
        } catch (error) {
          root.touch = false;
        }
      }
    }
  }

  MouseArea {
    anchors.fill: parent
    cursorShape: Qt.PointingHandCursor
    onClicked: Quickshell.execDetached(["solanum-look", "keyboard", "toggle"])
  }
}
