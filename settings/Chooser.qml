import QtQuick
import qs.common

Item {
  id: chooser

  property var row: null
  property int index: 0
  property var picked: []
  readonly property var choices: row ? (row.multiple ? [
        {
          value: "\n",
          label: "Apply the ticked ones"
        }
      ].concat(row.choices || []) : row.choices || []) : []

  signal closed

  function open(chosen) {
    row = chosen;
    picked = (chosen.value || "").split(" ").filter(word => word);
    index = Math.max(0, choices.findIndex(entry => entry.value === chosen.value));
    visible = true;
  }

  function close() {
    visible = false;
    row = null;
    closed();
  }

  function pick(entry) {
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

  function key(event) {
    if (event.key === Qt.Key_Escape)
      close();
    else if (event.key === Qt.Key_Up)
      index = Math.max(0, index - 1);
    else if (event.key === Qt.Key_Down || event.key === Qt.Key_Tab)
      index = Math.min(choices.length - 1, index + 1);
    else if (event.key === Qt.Key_Return || event.key === Qt.Key_Enter || event.key === Qt.Key_Space)
      pick(choices[index]);
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
      anchors.right: parent.right
      anchors.top: parent.top
      anchors.margins: 16
      text: chooser.row ? chooser.row.label : ""
      font.bold: true
      elide: Text.ElideRight
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
        icon: chooser.row && chooser.row.multiple && modelData.value !== "\n" ? (chooser.picked.includes(modelData.value) ? "\uf14a" : "\uf096") : modelData.value === (chooser.row ? chooser.row.value : "") ? "\uf00c" : " "
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
