import QtQuick
import QtQuick.Layouts
import Quickshell
import qs.common

FocusScope {
  id: page

  readonly property var waiting: Store.waiting
  readonly property bool working: Store.job === "update" && (Store.running || Store.task.code !== -1)
  readonly property var names: ({
      system: "Debian",
      apps: "Apps",
      solanum: "SolanumOS"
    })
  readonly property var words: ({
      ok: "PASS",
      fail: "FAIL",
      busy: "...",
      skip: "--"
    })
  readonly property var groups: waiting ? [
    {
      label: "Debian",
      list: waiting.debian,
      empty: "Up to date"
    },
    {
      label: "Apps",
      list: waiting.apps,
      empty: waiting.tools.length ? "Checked when the update runs: " + waiting.tools.join(", ") : "None to update"
    },
    {
      label: "SolanumOS",
      list: waiting.solanum,
      empty: "Up to date"
    }
  ].concat(waiting.firmware.length ? [
      {
        label: "Firmware",
        list: waiting.firmware,
        empty: ""
      }
    ] : []) : []
  readonly property var entries: groups.reduce((all, group) => all.concat([
      {
        heading: group.label,
        note: group.list.length ? group.list.length + " waiting" : group.empty
      }
    ], group.list.map(item => ({
        name: item.name,
        change: item.to ? (item.from ? item.from + " \u2192 " : "") + item.to : ""
      }))), [])
  readonly property var steps: stepped(Store.task.rows.count, Store.running, Store.task.code, Store.part)
  readonly property bool restart: !!waiting && waiting.restart && !Store.running
  readonly property var actions: [everything, debian, apps, solanum, firmware, restarting].filter(item => item.visible)
  readonly property var hints: working ? [] : [[["Left", "Right"], "pick"], ["Enter", "press"], [["Up", "Down"], "scroll"]]

  function begun(id, text) {
    if (id === "system")
      return text === "Updating Debian";
    if (id === "apps")
      return /^Updating (Flatpak apps|mise's tools|Rust)$/.test(text);
    return text.startsWith("Updating /");
  }

  function stepped() {
    const wanted = Store.part ? [Store.part] : ["system", "apps", "solanum"];
    const texts = [];
    for (let at = 0; at < Store.task.rows.count; at++)
      texts.push(Store.task.rows.get(at).text);
    const started = wanted.map(id => texts.some(text => begun(id, text)));
    const last = started.lastIndexOf(true);
    return wanted.map((id, at) => ({
          label: names[id],
          kind: !started[at] ? "skip" : at < last ? "ok" : Store.running ? "busy" : Store.task.code === 0 ? "ok" : "fail"
        }));
  }

  function enter() {
    if (working)
      view.forceActiveFocus();
    else if (actions.length)
      actions[0].forceActiveFocus();
    else
      page.forceActiveFocus();
  }

  function sideways(step) {
    const at = actions.findIndex(item => item.activeFocus);
    actions[Math.max(0, Math.min(actions.length - 1, at + step))].forceActiveFocus();
    return true;
  }

  function scroll(key) {
    const line = Theme.size(28);
    const steps = new Map([[Qt.Key_Up, -line], [Qt.Key_Down, line], [Qt.Key_PageUp, line - list.height], [Qt.Key_PageDown, list.height - line]]);
    if (!steps.has(key))
      return false;
    list.contentY = Math.max(list.originY, Math.min(list.contentY + steps.get(key), list.originY + list.contentHeight - list.height));
    return true;
  }

  function key(event) {
    if (working)
      return false;
    if ((event.key === Qt.Key_Left || event.key === Qt.Key_Right) && actions.length)
      return sideways(event.key === Qt.Key_Left ? -1 : 1);
    return scroll(event.key);
  }

  function run(args) {
    Store.start("update", args);
    Qt.callLater(enter);
  }

  Keys.onPressed: event => event.accepted = page.key(event)
  onWorkingChanged: Qt.callLater(enter)

  ColumnLayout {
    anchors.fill: parent
    spacing: 16
    visible: !page.working

    Label {
      Layout.fillWidth: true
      visible: !page.waiting
      text: "Looking for updates"
      color: Theme.subtext
    }

    ListView {
      id: list

      Layout.fillWidth: true
      Layout.fillHeight: true
      clip: true
      boundsBehavior: Flickable.StopAtBounds
      model: page.entries

      delegate: Item {
        id: entry

        required property var modelData
        required property int index

        width: ListView.view.width
        height: modelData.heading ? Theme.size(20) + (index > 0 ? 24 : 0) : Theme.size(26)

        Heading {
          anchors.left: parent.left
          anchors.right: parent.right
          anchors.bottom: parent.bottom
          visible: !!entry.modelData.heading
          text: entry.modelData.heading || ""
          note: entry.modelData.note || ""
        }

        Label {
          anchors.left: parent.left
          anchors.right: change.left
          anchors.rightMargin: 16
          anchors.verticalCenter: parent.verticalCenter
          visible: !entry.modelData.heading
          text: entry.modelData.name || ""
          elide: Text.ElideRight
        }

        Label {
          id: change

          anchors.right: parent.right
          anchors.verticalCenter: parent.verticalCenter
          visible: !entry.modelData.heading
          text: entry.modelData.change || ""
          color: Theme.subtext
          font.features: ({
              tnum: 1
            })
        }
      }
    }

    Label {
      Layout.fillWidth: true
      visible: !!page.waiting
      text: page.waiting && page.waiting.busy ? "Another Solanum run is going. Updating waits until it ends." : (page.waiting && page.waiting.last ? "Last updated " + Qt.formatDateTime(new Date(page.waiting.last), "d MMMM yyyy 'at' HH:mm") : "No update recorded yet") + ". A snapshot is taken first, so you can go back."
      color: Theme.subtext
      wrapMode: Text.Wrap
    }

    RowLayout {
      Layout.fillWidth: true
      spacing: 16

      Button {
        id: everything

        visible: !!page.waiting && !page.waiting.busy
        label: "Update everything"
        primary: true
        onClicked: page.run(["update"])
      }

      Button {
        id: debian

        visible: everything.visible
        label: "Debian only"
        onClicked: page.run(["update", "system"])
      }

      Button {
        id: apps

        visible: everything.visible
        label: "Apps only"
        onClicked: page.run(["update", "apps"])
      }

      Button {
        id: solanum

        visible: everything.visible
        label: "SolanumOS only"
        onClicked: page.run(["update", "solanum"])
      }

      Button {
        id: firmware

        visible: !!page.waiting && page.waiting.firmware.length > 0
        label: "Update firmware"
        onClicked: Quickshell.execDetached(["solanum-terminal", "--wait", "sh", "-c", "fwupdmgr refresh --force; fwupdmgr update"])
      }

      Button {
        id: restarting

        visible: page.restart
        label: "Restart now"
        onClicked: Quickshell.execDetached(["systemctl", "reboot"])
      }

      Item {
        Layout.fillWidth: true
      }

      Button {
        visible: !!page.waiting && page.waiting.busy
        label: "Look again"
        onClicked: Store.load()
      }
    }
  }

  ColumnLayout {
    anchors.fill: parent
    spacing: 16
    visible: page.working

    RowLayout {
      Layout.fillWidth: true
      spacing: 24

      Repeater {
        model: page.steps

        RowLayout {
          id: step

          required property var modelData

          spacing: 8

          Rectangle {
            Layout.preferredWidth: shown.implicitWidth + 12
            Layout.preferredHeight: Theme.size(24)
            color: step.modelData.kind === "fail" ? Theme.text : "transparent"

            RowLayout {
              id: shown

              anchors.centerIn: parent
              spacing: 8

              Mark {
                kind: step.modelData.kind
                ink: step.modelData.kind === "fail" ? Theme.base : Theme.text
              }

              Label {
                text: page.words[step.modelData.kind] + "  " + step.modelData.label
                color: step.modelData.kind === "fail" ? Theme.base : Theme.text
                font.family: Theme.display
                font.bold: Theme.displayBold
              }
            }
          }
        }
      }

      Item {
        Layout.fillWidth: true
      }

      Button {
        visible: page.restart && Store.task.code === 0
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
      title: Store.part ? "Updating " + page.names[Store.part] : "Updating everything"
      onClosed: Store.finish()
    }
  }
}
