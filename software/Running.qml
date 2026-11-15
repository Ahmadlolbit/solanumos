import QtQuick
import qs.common

FocusScope {
  id: running

  signal closed

  Label {
    id: none

    anchors.left: parent.left
    anchors.right: parent.right
    anchors.top: parent.top
    anchors.leftMargin: 16
    anchors.rightMargin: 16
    visible: !Store.job
    text: "Nothing has run yet. The changes you apply show here, one after another."
    color: Theme.subtext
    wrapMode: Text.Wrap
  }

  TaskView {
    id: view

    anchors.left: parent.left
    anchors.right: parent.right
    anchors.top: parent.top
    anchors.bottom: parent.bottom
    visible: !!Store.job
    focus: true
    task: Store.task
    title: Store.job ? Store.job.label + (Store.queue.length > 0 ? ", then " + Store.queue.length + " more" : "") : ""
    onClosed: running.closed()
  }
}
