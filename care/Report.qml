import QtQuick
import QtQuick.Layouts
import Quickshell
import qs.common

FocusScope {
  id: page

  readonly property bool working: Store.job === "collect" && (Store.running || Store.task.code !== -1)
  readonly property string made: working && !Store.running && Store.task.code === 0 ? saved(Store.task.rows.count) : ""
  readonly property string shown: made || Store.report
  readonly property var gathered: ["The machine, Debian and the kernel, the devices, the network's state and the btrfs mounts", "The versions of Hyprland, Quickshell, the kernels, the firmware and the drivers", "This start's warnings from the system journal and the kernel", "The screens, the workspaces, the theme, the desktop's programs and the bar's log", "Solanum's logs, the setup's logs and report, your configuration and how each part of the setup last ended", "The desktop session's log and Hyprland's logs"]
  readonly property var hints: working ? [] : [[["Left", "Right"], "pick"], ["Enter", "press"]]
  readonly property var actions: [make, copy].filter(item => item.visible)

  function saved() {
    for (let at = Store.task.rows.count - 1; at >= 0; at--) {
      const text = Store.task.rows.get(at).text;
      if (text.startsWith("Saved "))
        return text.slice(6).trim();
    }
    return "";
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

  function makeOne() {
    if (Store.running)
      return;
    Store.start("collect", ["collect"]);
    Qt.callLater(enter);
  }

  Keys.onPressed: event => event.accepted = !working && (event.key === Qt.Key_Left || event.key === Qt.Key_Right) && sideways(event.key === Qt.Key_Left ? -1 : 1)
  onWorkingChanged: Qt.callLater(enter)
  onMadeChanged: {
    if (made)
      Store.report = made;
  }

  ColumnLayout {
    anchors.left: parent.left
    anchors.right: parent.right
    anchors.top: parent.top
    spacing: 16
    visible: !page.working

    Label {
      Layout.fillWidth: true
      text: "A report puts what helps find a problem in one file in your home folder. Nothing is sent anywhere: you attach the file to an issue yourself. It asks for your password, since some logs are only root's to read."
      wrapMode: Text.Wrap
    }

    Heading {
      Layout.fillWidth: true
      text: "What it gathers"
    }

    Repeater {
      model: page.gathered

      Label {
        required property string modelData

        Layout.fillWidth: true
        Layout.leftMargin: 16
        text: "–  " + modelData
        wrapMode: Text.Wrap
      }
    }

    Label {
      Layout.fillWidth: true
      text: "Network addresses are masked, and window titles and serial numbers are hidden."
      color: Theme.subtext
      wrapMode: Text.Wrap
    }

    Label {
      Layout.fillWidth: true
      visible: page.shown !== ""
      text: "The report: " + page.shown
      wrapMode: Text.WrapAnywhere
    }

    RowLayout {
      spacing: 16

      Button {
        id: make

        visible: !Store.running
        label: "Make a report"
        primary: true
        onClicked: page.makeOne()
      }

      Button {
        id: copy

        visible: page.shown !== ""
        label: "Copy the path"
        onClicked: Quickshell.execDetached(["wl-copy", "--", page.shown])
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
      visible: page.made !== ""

      Label {
        Layout.fillWidth: true
        text: "Saved " + page.made
        wrapMode: Text.WrapAnywhere
      }

      Button {
        label: "Copy the path"
        onClicked: Quickshell.execDetached(["wl-copy", "--", page.made])
      }
    }

    TaskView {
      id: view

      Layout.fillWidth: true
      Layout.fillHeight: true
      task: Store.task
      held: Store.holding
      stopQuestion: "Stop the report?"
      title: "Making a report"
      onClosed: Store.finish()
    }
  }
}
