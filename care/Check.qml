import QtQuick
import QtQuick.Layouts
import qs.common

FocusScope {
  id: page

  property bool fixAll: false
  readonly property bool working: Store.job === "doctor" && (Store.running || Store.task.code !== -1)
  readonly property bool fixing: working && Store.task.question !== null && Store.task.question.kind === "choose"
  readonly property bool going: !!Store.waiting && Store.waiting.busy
  readonly property var hints: working ? [] : start.visible ? [["Enter", "check"]] : []

  function enter() {
    if (working)
      view.forceActiveFocus();
    else if (start.visible)
      start.forceActiveFocus();
    else
      page.forceActiveFocus();
  }

  function check() {
    fixAll = false;
    Store.start("doctor", ["doctor"]);
    Qt.callLater(enter);
  }

  function everyFix() {
    fixAll = true;
    Store.task.answer("Yes");
  }

  onWorkingChanged: Qt.callLater(enter)

  function settle() {
    const asked = Store.task.question;
    if (page.fixAll && !Store.holding && Store.job === "doctor" && asked && asked.kind === "choose")
      Store.task.answer("Yes");
  }

  Connections {
    target: Store.task

    function onQuestionChanged() {
      Qt.callLater(page.settle);
    }
  }

  Connections {
    target: Store

    function onHoldingChanged() {
      Qt.callLater(page.settle);
    }
  }

  ColumnLayout {
    anchors.left: parent.left
    anchors.right: parent.right
    anchors.top: parent.top
    spacing: 16
    visible: !page.working

    Label {
      Layout.fillWidth: true
      text: "Looks at the desktop, the storage, the hardware and packages, and Solanum itself, each with PASS, WARN or FAIL. It asks for your password, then offers each safe fix it finds, one at a time."
      wrapMode: Text.Wrap
    }

    Label {
      Layout.fillWidth: true
      visible: page.going || Store.running
      text: Store.running ? "An update is running. The check can start once it ends." : "Another Solanum run is going. The check runs, and its fixes wait until that run ends."
      color: Theme.subtext
      wrapMode: Text.Wrap
    }

    Button {
      id: start

      visible: !Store.running
      label: "Check the system"
      primary: true
      onClicked: page.check()
    }
  }

  ColumnLayout {
    anchors.fill: parent
    spacing: 16
    visible: page.working

    TaskView {
      id: view

      Layout.fillWidth: true
      Layout.fillHeight: true
      task: Store.task
      title: "System check"
      onClosed: Store.finish()
    }

    RowLayout {
      Layout.fillWidth: true
      spacing: 16
      visible: page.fixing

      Label {
        Layout.fillWidth: true
        text: "Each fix is offered on its own. Fix all says yes to this one and the rest."
        color: Theme.subtext
        wrapMode: Text.Wrap
      }

      Button {
        label: "Fix all"
        onClicked: page.everyFix()
      }
    }
  }
}
