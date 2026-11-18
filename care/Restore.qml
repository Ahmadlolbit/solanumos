import QtQuick
import QtQuick.Layouts
import Quickshell
import qs.common
import "." as Care

FocusScope {
  id: page

  property int index: 0
  readonly property var points: Store.points || []
  readonly property var current: points.length ? points[Math.min(index, points.length - 1)] : null
  readonly property var result: current ? Store.compared[current.number] || null : null
  readonly property bool working: Store.job === "restore" && (Store.running || Store.task.code !== -1)
  readonly property bool ended: working && !Store.running
  readonly property var said: ended && Store.task.code !== 0 ? lastLines(Store.task.log.count, Store.task.reason) : []
  readonly property var hints: working ? [] : points.length ? [[["Up", "Down"], "point"], ["Enter", "go back to it"]] : []

  function grouped(count) {
    return String(count).replace(/\B(?=(\d{3})+(?!\d))/g, ",");
  }

  function when(point) {
    return Qt.formatDateTime(new Date(point.time), "ddd d MMM yyyy, HH:mm");
  }

  function lastLines() {
    const lines = [];
    for (let at = Math.max(0, Store.task.log.count - 8); at < Store.task.log.count; at++)
      lines.push(Store.task.log.get(at).text);
    for (const line of Store.task.reason.split("\n"))
      if (line.trim() && !lines.includes(line))
        lines.push(line);
    return lines.filter(line => line.trim()).slice(-10);
  }

  function enter() {
    if (working)
      view.forceActiveFocus();
    else
      page.forceActiveFocus();
  }

  function pick(at) {
    index = Math.max(0, Math.min(points.length - 1, at));
    if (current) {
      Store.point = current.number;
      Store.compare(current.number);
    }
  }

  function settle() {
    const found = points.findIndex(point => point.number === Store.point);
    pick(found >= 0 ? found : 0);
  }

  function goBack() {
    if (!current || Store.running)
      return;
    Store.start("restore", ["restore", current.number]);
    Qt.callLater(enter);
  }

  function key(event) {
    if (working || !points.length)
      return false;
    const steps = new Map([[Qt.Key_Up, -1], [Qt.Key_Down, 1], [Qt.Key_PageUp, -8], [Qt.Key_PageDown, 8]]);
    if (steps.has(event.key)) {
      pick(index + steps.get(event.key));
      return true;
    }
    if (event.key !== Qt.Key_Return && event.key !== Qt.Key_Enter)
      return false;
    goBack();
    return true;
  }

  Keys.onPressed: event => event.accepted = page.key(event)
  onWorkingChanged: Qt.callLater(enter)
  onPointsChanged: Qt.callLater(settle)
  Component.onCompleted: Qt.callLater(settle)

  RowLayout {
    anchors.fill: parent
    spacing: 24
    visible: !page.working

    ColumnLayout {
      Layout.preferredWidth: Math.round(parent.width * 0.42)
      Layout.fillHeight: true
      spacing: 8

      Label {
        Layout.fillWidth: true
        visible: !Store.points || page.points.length === 0
        text: !Store.points ? "Looking for restore points" : Store.pointsWhy ? "The restore points can't be read: " + Store.pointsWhy : "No restore points yet. One is taken before every update, and one every day."
        color: Theme.subtext
        wrapMode: Text.Wrap
      }

      ListView {
        id: list

        Layout.fillWidth: true
        Layout.fillHeight: true
        clip: true
        spacing: 4
        model: page.points
        currentIndex: page.index
        highlightMoveDuration: 0
        boundsBehavior: Flickable.StopAtBounds
        onCurrentIndexChanged: positionViewAtIndex(currentIndex, ListView.Contain)

        delegate: Care.ListRow {
          required property var modelData
          required property int index

          width: ListView.view.width
          title: "#" + modelData.number + "  " + page.when(modelData)
          detail: modelData.description
          current: index === page.index
          framed: page.activeFocus
          onPicked: {
            page.pick(index);
            page.forceActiveFocus();
          }
        }
      }
    }

    ColumnLayout {
      Layout.fillWidth: true
      Layout.preferredWidth: 1
      Layout.fillHeight: true
      Layout.alignment: Qt.AlignTop
      spacing: 16
      visible: !!page.current

      Heading {
        Layout.fillWidth: true
        text: "Since then"
        note: page.current ? "#" + page.current.number : ""
      }

      RowLayout {
        Layout.fillWidth: true
        spacing: 8
        visible: !page.result

        Care.Mark {
          kind: "busy"
        }

        Label {
          text: "RUNNING"
          font.family: Theme.display
          font.bold: Theme.displayBold
        }

        Label {
          Layout.fillWidth: true
          text: Store.comparing && page.current && Store.comparing !== page.current.number ? "Waiting to compare. Another point is being compared first." : "Comparing with today. This can take a minute, and you can look at other points meanwhile."
          color: Theme.subtext
          wrapMode: Text.Wrap
        }
      }

      RowLayout {
        Layout.fillWidth: true
        spacing: 8
        visible: !!page.result && !!page.result.failed

        Care.Mark {
          kind: "fail"
        }

        Label {
          Layout.fillWidth: true
          text: "FAIL  " + (page.result && page.result.failed ? page.result.failed : "")
          wrapMode: Text.Wrap
        }
      }

      Label {
        Layout.fillWidth: true
        visible: !!page.result && !page.result.failed
        text: page.result && !page.result.failed ? page.grouped(page.result.added) + " added, " + page.grouped(page.result.removed) + " removed, " + page.grouped(page.result.changed) + " changed" : ""
        font.bold: true
        font.features: ({
            tnum: 1
          })
      }

      Repeater {
        model: page.result && page.result.folders ? page.result.folders : []

        RowLayout {
          id: place

          required property var modelData

          Layout.fillWidth: true
          spacing: 16

          Label {
            Layout.preferredWidth: Theme.size(64)
            horizontalAlignment: Text.AlignRight
            text: page.grouped(place.modelData.count)
            color: Theme.subtext
            font.features: ({
                tnum: 1
              })
          }

          Label {
            Layout.fillWidth: true
            text: place.modelData.folder
            elide: Text.ElideMiddle
          }
        }
      }

      Heading {
        Layout.fillWidth: true
        visible: !!page.result && !!page.result.kept
        text: "What stays"
      }

      Label {
        Layout.fillWidth: true
        visible: !!page.result && !!page.result.kept
        text: page.result && page.result.kept ? page.result.kept : ""
        wrapMode: Text.Wrap
      }

      RowLayout {
        Layout.fillWidth: true
        spacing: 8
        visible: !!page.result && !!page.result.kernel

        Care.Mark {
          Layout.alignment: Qt.AlignTop
          kind: "warn"
        }

        Label {
          Layout.fillWidth: true
          text: "WARN  " + (page.result && page.result.kernel ? page.result.kernel : "")
          wrapMode: Text.Wrap
        }
      }

      Item {
        Layout.fillHeight: true
      }

      Label {
        Layout.fillWidth: true
        visible: !!Store.waiting && Store.waiting.busy
        text: "Another Solanum run is going. Going back waits until it ends."
        color: Theme.subtext
        wrapMode: Text.Wrap
      }

      Button {
        visible: !!page.current && !Store.running
        label: "Go back to this"
        primary: true
        onClicked: page.goBack()
      }
    }
  }

  ColumnLayout {
    anchors.fill: parent
    spacing: 16
    visible: page.working

    RowLayout {
      Layout.fillWidth: true
      spacing: 16

      Label {
        Layout.fillWidth: true
        visible: page.ended && Store.task.code === 0
        text: "Done. The computer starts from the restore point at the next restart."
        wrapMode: Text.Wrap
      }

      Item {
        Layout.fillWidth: true
        visible: !(page.ended && Store.task.code === 0)
      }

      Button {
        visible: page.ended && Store.task.code === 0
        label: "Restart now"
        primary: true
        onClicked: Quickshell.execDetached(["systemctl", "reboot"])
      }
    }

    TaskView {
      id: view

      Layout.fillWidth: true
      Layout.fillHeight: true
      task: Store.task
      held: Store.holding
      stopQuestion: "Stop going back?"
      title: Store.part ? "Going back to restore point #" + Store.part : "Going back"
      onClosed: Store.finish()
    }

    Rectangle {
      Layout.fillWidth: true
      Layout.preferredHeight: last.implicitHeight + 32
      visible: page.said.length > 0
      color: Theme.surface
      radius: Theme.radius
      border.width: 1
      border.color: Theme.overlay

      ColumnLayout {
        id: last

        anchors.fill: parent
        anchors.margins: 16
        spacing: 4

        Heading {
          Layout.fillWidth: true
          text: "What it said last"
          note: "Exit code " + Store.task.code
        }

        Repeater {
          model: page.said

          Label {
            required property string modelData

            Layout.fillWidth: true
            text: modelData
            wrapMode: Text.Wrap
          }
        }
      }
    }
  }
}
