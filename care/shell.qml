import QtQuick
import Quickshell
import Quickshell.Io
import qs.common

ShellRoot {
  id: root

  IpcHandler {
    target: "care"

    function open(page: string): void {
      if (page)
        Store.show(page);
      Store.load();
      window.visible = true;
      frame.takeKeys();
    }

    function close(): void {
      frame.leave();
    }
  }

  FloatingWindow {
    id: window

    title: "Care"
    visible: true
    implicitWidth: Number(Quickshell.env("SOLANUM_CARE_WIDTH")) || 1100
    implicitHeight: Number(Quickshell.env("SOLANUM_CARE_HEIGHT")) || 720
    minimumSize: Qt.size(800, 560)
    color: Theme.base
    onVisibleChanged: {
      if (visible)
        return;
      if (!Store.running)
        Qt.quit();
      else
        Qt.callLater(() => {
          visible = true;
          frame.ask();
        });
    }

    Frame {
      id: frame

      anchors.fill: parent
      onDone: Qt.quit()
    }
  }

  Component.onCompleted: Store.load()
}
