import QtQuick
import QtQuick.Layouts
import qs.common

FocusScope {
  id: page

  readonly property var plan: Store.plan
  readonly property bool working: Store.job === "cleanup" && (Store.running || Store.task.code !== -1)
  readonly property bool known: !!plan && plan.known
  readonly property int old: plan ? plan.old.length : 0
  readonly property bool wanted: known && plan.use >= plan.goal && old > 0 && !(Store.waiting && Store.waiting.busy)
  readonly property var hints: working ? [] : wanted ? [["Enter", "delete old snapshots"]] : []

  function size(bytes) {
    const units = ["B", "KiB", "MiB", "GiB", "TiB"];
    let at = 0;
    let value = bytes;
    while (value >= 1024 && at < units.length - 1) {
      value /= 1024;
      at++;
    }
    return (at ? value.toFixed(1) : String(value)) + " " + units[at];
  }

  function told() {
    if (!plan)
      return "Looking at the disk";
    if (!known)
      return "How full / is can't be told.";
    if (plan.unread)
      return "/ is " + plan.use + "% full. The snapshots can't be read.";
    return "/ is " + plan.use + "% full." + (plan.use >= plan.goal ? " Cleanup deletes old snapshots until it's under " + plan.goal + "%." : " Cleanup starts when it's " + plan.goal + "% full.");
  }

  function oldWords() {
    if (!plan || plan.unread || !known)
      return "";
    if (!old)
      return "No old automatic snapshots can go. The newest three and the ones with your own description always stay.";
    return (old === 1 ? "1 old automatic snapshot" : old + " old automatic snapshots") + " of " + plan.snapshots + " can go, oldest first. The newest three and the ones with your own description stay." + (plan.heldKnown ? " They hold " + size(plan.held) + "." : "");
  }

  function enter() {
    if (working)
      view.forceActiveFocus();
    else if (wanted)
      clean.forceActiveFocus();
    else
      page.forceActiveFocus();
  }

  function cleanUp() {
    if (Store.running)
      return;
    Store.start("cleanup", ["cleanup"]);
    Qt.callLater(enter);
  }

  onWorkingChanged: Qt.callLater(enter)
  onWantedChanged: Qt.callLater(enter)

  ColumnLayout {
    anchors.left: parent.left
    anchors.right: parent.right
    anchors.top: parent.top
    spacing: 16
    visible: !page.working

    Heading {
      Layout.fillWidth: true
      text: "The disk"
      note: page.known ? page.plan.use + "%" : ""
    }

    Rectangle {
      Layout.fillWidth: true
      Layout.preferredHeight: Theme.size(16)
      visible: page.known
      color: "transparent"
      border.width: 1
      border.color: Theme.text

      Rectangle {
        anchors.left: parent.left
        anchors.top: parent.top
        anchors.bottom: parent.bottom
        width: page.known ? Math.round(parent.width * Math.min(100, page.plan.use) / 100) : 0
        color: Theme.text
      }

      Rectangle {
        x: page.known ? Math.round(parent.width * page.plan.goal / 100) : 0
        anchors.top: parent.top
        anchors.bottom: parent.bottom
        anchors.topMargin: -4
        anchors.bottomMargin: -4
        width: 2
        color: Theme.accent
      }
    }

    Label {
      Layout.fillWidth: true
      text: page.told()
      wrapMode: Text.Wrap
    }

    Heading {
      Layout.fillWidth: true
      Layout.topMargin: 8
      visible: page.oldWords() !== ""
      text: "Old snapshots"
      note: page.old ? String(page.old) : ""
    }

    Label {
      Layout.fillWidth: true
      visible: text !== ""
      text: page.oldWords()
      wrapMode: Text.Wrap
    }

    Label {
      Layout.fillWidth: true
      visible: !!Store.waiting && Store.waiting.busy
      text: "Another Solanum run is going. Old snapshots can be deleted once it ends."
      color: Theme.subtext
      wrapMode: Text.Wrap
    }

    Button {
      id: clean

      visible: page.wanted && !Store.running
      label: "Delete old snapshots"
      primary: true
      onClicked: page.cleanUp()
    }
  }

  TaskView {
    id: view

    anchors.fill: parent
    visible: page.working
    task: Store.task
    held: Store.holding
    stopQuestion: "Stop deleting snapshots?"
    title: "Deleting old snapshots"
    onClosed: Store.finish()
  }
}
