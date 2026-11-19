import QtQuick
import qs.common

Item {
  id: chooser

  property var row: null
  property int index: 0
  property var picked: []
  property string filter: ""
  readonly property var hints: [[["Up", "Down"], "move"], ["Enter", row && row.multiple ? "tick" : "pick"], ["", "type to filter"], ["Esc", filter ? "clear" : "close"]]
  readonly property var matching: (row ? row.choices || [] : []).filter(entry => entry.label.toLowerCase().includes(filter.toLowerCase()))
  readonly property var choices: row ? (row.multiple ? [
        {
          value: "\n",
          label: "Apply the ticked ones"
        }
      ].concat(matching) : matching) : []

  signal closed

  function open(chosen) {
    filter = "";
    row = chosen;
    picked = (chosen.value || "").split(" ").filter(word => word);
    index = Math.max(0, choices.findIndex(entry => entry.value === chosen.value));
    visible = true;
  }

  function close() {
    visible = false;
    filter = "";
    row = null;
    closed();
  }

  function pick(entry) {
    if (!entry)
      return;
    if (!row.multiple) {
      Store.set(row, entry.value);
      close();
    } else if (entry.value === "\n") {
      Store.set(row, picked.join(" "));
      close();
    } else {
      picked = picked.includes(entry.value) ? picked.filter(word => word !== entry.value) : picked.concat([entry.value]);
    }
  }

  function typed(event) {
    if (event.key === Qt.Key_Backspace && filter)
      filter = filter.slice(0, -1);
    else if (Keyboard.joins(filter, event.text))
      filter += event.text;
    else
      return;
    index = Math.max(0, choices.findIndex(entry => entry.value !== "\n"));
  }

  function key(event) {
    const moves = {
      [Qt.Key_Up]: index - 1,
      [Qt.Key_Down]: index + 1,
      [Qt.Key_Tab]: index + 1,
      [Qt.Key_Home]: 0,
      [Qt.Key_End]: choices.length - 1
    };
    if (event.key === Qt.Key_Escape && filter)
      filter = "";
    else if (event.key === Qt.Key_Escape)
      close();
    else if (moves[event.key] !== undefined)
      index = Math.max(0, Math.min(choices.length - 1, moves[event.key]));
    else if ([Qt.Key_Return, Qt.Key_Enter].includes(event.key) || event.key === Qt.Key_Space && !filter)
      pick(choices[index]);
    else
      typed(event);
    return true;
  }

  visible: false

  Rectangle {
    anchors.fill: parent
    color: Qt.alpha(Theme.base, 0.7)

    MouseArea {
      anchors.fill: parent
      onClicked: chooser.close()
    }
  }

  Rectangle {
    anchors.centerIn: parent
    width: Math.min(parent.width - 64, Theme.size(460))
    height: Math.min(parent.height - 64, heading.height + list.contentHeight + 40)
    color: Theme.base
    radius: Theme.radius
    border.width: 2
    border.color: Theme.accent

    MouseArea {
      anchors.fill: parent
    }

    Label {
      id: heading

      anchors.left: parent.left
      anchors.right: sought.left
      anchors.top: parent.top
      anchors.margins: 16
      text: chooser.row ? chooser.row.label : ""
      font.bold: true
      elide: Text.ElideRight
    }

    Label {
      id: sought

      anchors.right: parent.right
      anchors.verticalCenter: heading.verticalCenter
      anchors.rightMargin: 16
      width: Math.min(implicitWidth, parent.width / 2)
      text: chooser.filter ? "\uf002  " + chooser.filter : "Type to filter"
      color: chooser.filter ? Theme.text : Theme.subtext
      elide: Text.ElideLeft
    }

    ListView {
      id: list

      anchors.left: parent.left
      anchors.right: parent.right
      anchors.top: heading.bottom
      anchors.bottom: parent.bottom
      anchors.margins: 10
      clip: true
      model: chooser.choices
      currentIndex: chooser.index
      highlightMoveDuration: 0
      boundsBehavior: Flickable.StopAtBounds
      onCurrentIndexChanged: positionViewAtIndex(currentIndex, ListView.Contain)

      delegate: Choice {
        required property var modelData
        required property int index

        width: ListView.view.width
        icon: chooser.row && chooser.row.multiple && modelData.value !== "\n" ? (chooser.picked.includes(modelData.value) ? "\uf14a" : "\uf096") : modelData.value === (chooser.row ? chooser.row.value : "") ? "\u25a0" : " "
        label: modelData.label
        selected: index === chooser.index
        onClicked: {
          chooser.index = index;
          chooser.pick(modelData);
        }
      }
    }
  }
}
