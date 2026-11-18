import QtQuick
import QtQuick.Layouts
import qs.common
import "." as Care

FocusScope {
  id: page

  property int index: 0
  readonly property var modules: Store.undos || []
  readonly property var current: modules.length ? modules[Math.min(index, modules.length - 1)] : null
  readonly property bool working: Store.job === "undo" && (Store.running || Store.task.code !== -1)
  readonly property var states: ({
      ready: "Set up",
      undone: "Undone",
      unknown: "Not run",
      unread: "Unread"
    })
  readonly property var hints: working ? [] : current && current.state === "ready" ? [[["Up", "Down"], "part"], ["Enter", "undo it"]] : [[["Up", "Down"], "part"]]

  function named(noun, names) {
    if (!names.length)
      return "";
    const shown = names.slice(0, 8).join(" ");
    const word = names.length === 1 ? noun.replace(/s$/, "") : noun;
    return names.length + " " + word + ": " + shown + (names.length > 8 ? " and " + (names.length - 8) + " more" : "");
  }

  function takes(item) {
    if (!item)
      return [];
    if (item.state === "undone")
      return ["It's undone. The setup leaves it out until you run it again: sudo solanum install --only " + item.id];
    if (item.state === "unknown")
      return ["There's no record of it yet. Run the setup once, then it can be undone."];
    if (item.state === "unread")
      return [item.problem];
    const lines = [named("packages", item.packages), named("services", item.services), named("files", item.files), named("unpacked", item.unpacked)].filter(line => line);
    return lines.length ? lines : ["Nothing left to take back but its record."];
  }

  function enter() {
    if (working)
      view.forceActiveFocus();
    else
      page.forceActiveFocus();
  }

  function undo() {
    if (!current || current.state !== "ready" || Store.running)
      return;
    Store.start("undo", ["undo", current.id]);
    Qt.callLater(enter);
  }

  function key(event) {
    if (working || !modules.length)
      return false;
    if (event.key === Qt.Key_Up || event.key === Qt.Key_Down) {
      index = Math.max(0, Math.min(modules.length - 1, index + (event.key === Qt.Key_Up ? -1 : 1)));
      return true;
    }
    if (event.key !== Qt.Key_Return && event.key !== Qt.Key_Enter)
      return false;
    undo();
    return true;
  }

  Keys.onPressed: event => event.accepted = page.key(event)
  onWorkingChanged: Qt.callLater(enter)

  RowLayout {
    anchors.fill: parent
    spacing: 24
    visible: !page.working

    ColumnLayout {
      Layout.preferredWidth: Math.round(parent.width * 0.42)
      Layout.fillHeight: true
      spacing: 4

      Label {
        Layout.fillWidth: true
        visible: !Store.undos
        text: "Reading what each part set up"
        color: Theme.subtext
      }

      Repeater {
        model: page.modules

        Care.ListRow {
          required property var modelData
          required property int index

          Layout.fillWidth: true
          title: modelData.id + "  " + modelData.title
          word: page.states[modelData.state] || ""
          current: index === page.index
          framed: page.activeFocus
          onPicked: {
            page.index = index;
            page.forceActiveFocus();
          }
        }
      }

      Item {
        Layout.fillHeight: true
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
        text: page.current && page.current.state === "ready" ? "Undoing it takes back" : "Undo"
        note: page.current ? page.current.id : ""
      }

      Repeater {
        model: page.takes(page.current)

        Label {
          required property string modelData

          Layout.fillWidth: true
          text: modelData
          wrapMode: Text.Wrap
        }
      }

      Label {
        Layout.fillWidth: true
        visible: !!page.current && page.current.state === "ready"
        text: "Packages and files another part also uses stay, and so do packages you asked apt for yourself and files you've changed since. A restore point is taken first."
        color: Theme.subtext
        wrapMode: Text.Wrap
      }

      Item {
        Layout.fillHeight: true
      }

      Button {
        visible: !!page.current && page.current.state === "ready" && !Store.running
        label: "Undo"
        primary: true
        onClicked: page.undo()
      }
    }
  }

  TaskView {
    id: view

    anchors.fill: parent
    visible: page.working
    task: Store.task
    held: Store.holding
    stopQuestion: "Stop the undo?"
    title: "Undoing module " + Store.part
    onClosed: Store.finish()
  }
}
