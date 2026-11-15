pragma Singleton

import QtQuick
import Quickshell
import Quickshell.Io

Singleton {
  id: root

  property int count: 0

  function refresh() {
    counter.running = true;
  }

  function upgrade() {
    Quickshell.execDetached(["solanum", "care", "updates"]);
  }

  Process {
    id: counter

    command: ["sh", "-c", "LC_ALL=C LANGUAGE= apt list --upgradable 2>/dev/null | grep -c upgradable"]
    stdout: StdioCollector {
      onStreamFinished: root.count = parseInt(text.trim()) || 0
    }
  }

  Timer {
    interval: 1800000
    running: true
    repeat: true
    triggeredOnStart: true
    onTriggered: root.refresh()
  }

  IpcHandler {
    target: "updates"

    function refresh(): void {
      root.refresh();
    }
  }
}
