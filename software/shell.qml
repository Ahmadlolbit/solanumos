import QtQuick
import Quickshell
import Quickshell.Io
import qs.common

ShellRoot {
  id: root

  IpcHandler {
    target: "software"

    function open(page: string, item: string): void {
      Store.search = "";
      Store.show(page);
      Store.wanted = item;
      Store.load();
      frame.zone = "rows";
      frame.takeKeys();
    }

    function close(): void {
      if (Store.working)
        frame.askStop();
      else
        Qt.quit();
    }
  }

  FloatingWindow {
    id: window

    title: "Software"
    visible: true
    implicitWidth: Number(Quickshell.env("SOLANUM_SOFTWARE_WIDTH")) || 1100
    implicitHeight: Number(Quickshell.env("SOLANUM_SOFTWARE_HEIGHT")) || 720
    minimumSize: Qt.size(800, 560)
    color: Theme.base
    onVisibleChanged: {
      if (!visible && !Store.working)
        Qt.quit();
    }
    onClosed: {
      if (Store.working)
        Qt.callLater(() => {
          visible = false;
          visible = true;
          frame.askStop();
        });
    }

    Frame {
      id: frame

      anchors.fill: parent
      onDone: Qt.quit()
      Component.onCompleted: Store.load()
    }
  }
}
