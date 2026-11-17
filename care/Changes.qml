import QtQuick
import QtQuick.Layouts
import qs.common

FocusScope {
  id: page

  readonly property var notes: Store.notes
  readonly property var entries: notes ? notes.groups.reduce((all, group) => all.concat(group.heading ? [
        {
          heading: group.heading
        }
      ] : [], group.items.map(item => ({
          item: item
        }))), []) : []
  readonly property var hints: [[["Up", "Down"], "scroll"]]

  function enter() {
    page.forceActiveFocus();
  }

  function scroll(key) {
    const line = Theme.size(26);
    const steps = new Map([[Qt.Key_Up, -line], [Qt.Key_Down, line], [Qt.Key_PageUp, line - list.height], [Qt.Key_PageDown, list.height - line]]);
    if (!steps.has(key))
      return false;
    list.contentY = Math.max(list.originY, Math.min(list.contentY + steps.get(key), list.originY + list.contentHeight - list.height));
    return true;
  }

  Keys.onPressed: event => event.accepted = page.scroll(event.key)

  ColumnLayout {
    anchors.fill: parent
    spacing: 16

    Heading {
      Layout.fillWidth: true
      text: "The last update"
      note: page.notes && page.notes.date ? Qt.formatDateTime(new Date(page.notes.date), "d MMMM yyyy") : ""
    }

    Label {
      Layout.fillWidth: true
      visible: !!page.notes && page.entries.length === 0
      text: "No update notes yet. They're written when an update brings a new version of SolanumOS."
      color: Theme.subtext
      wrapMode: Text.Wrap
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
        height: text.implicitHeight + (modelData.heading && index > 0 ? 24 : 8)

        Label {
          id: text

          anchors.left: parent.left
          anchors.right: parent.right
          anchors.bottom: parent.bottom
          anchors.leftMargin: entry.modelData.heading ? 0 : 16
          text: entry.modelData.heading || "\u2013  " + entry.modelData.item
          font.bold: !!entry.modelData.heading
          wrapMode: Text.Wrap
        }
      }
    }
  }
}
