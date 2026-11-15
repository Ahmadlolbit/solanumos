import QtQuick
import Quickshell
import Quickshell.Io
import qs.common

ShellRoot {
  id: root

  IpcHandler {
    target: "settings"

    function open(page: string): void {
      if (page)
        Store.show(page);
      Store.load();
      frame.forceActiveFocus();
    }

    function close(): void {
      Qt.quit();
    }
  }

  FloatingWindow {
    id: window

    title: "Settings"
    visible: true
    implicitWidth: Number(Quickshell.env("SOLANUM_SETTINGS_WIDTH")) || 1100
    implicitHeight: Number(Quickshell.env("SOLANUM_SETTINGS_HEIGHT")) || 720
    minimumSize: Qt.size(800, 560)
    color: Theme.base
    onVisibleChanged: {
      if (!visible)
        Qt.quit();
    }

    Frame {
      id: frame

      anchors.fill: parent
      onDone: Qt.quit()
    }
  }
}
