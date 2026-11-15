import QtQuick
import QtQuick.Layouts
import qs.common

Rectangle {
  id: item

  required property var row
  property bool current: false
  property bool searching: false
  property bool grouped: false
  property bool editing: false
  readonly property bool computer: row.whose === "computer"
  readonly property var choice: (row.choices || []).find(entry => entry.value === row.value)
  readonly property string shownValue: row.multiple ? (row.value || "none").split(" ").filter(word => word).join(", ") : choice ? choice.label : row.value
  readonly property bool busy: Store.busy === row.name
  readonly property var marks: [computer && !grouped ? "\uf023 For everyone on this computer" : "", row.after === "restart" ? "\uf021 Applies after a restart" : row.after === "login" ? "\uf021 Applies when you log in again" : "", searching ? "\uf0c9 " + ((Store.pages.find(page => page.id === row.page) || {}).label || "") : ""].filter(mark => mark)

  signal chosen
  signal pointed

  function activate() {
    if (row.kind === "toggle")
      Store.set(row, row.value === "yes" ? "no" : "yes");
    else if (row.kind === "choice")
      chosen();
    else if (row.kind === "button")
      Store.launch(row.command);
    else if (row.kind === "text" || row.kind === "time")
      edit();
    else
      return false;
    return true;
  }

  function step(direction) {
    if (row.kind === "slider") {
      const next = Math.max(row.min, Math.min(row.max, Number(row.value) + direction * row.step));
      Store.set(row, Math.round(next * 100) / 100);
      return true;
    }
    if (row.kind !== "choice" || row.multiple || !(row.choices || []).length)
      return false;
    const at = row.choices.findIndex(entry => entry.value === row.value);
    Store.set(row, row.choices[(at + direction + row.choices.length) % row.choices.length].value);
    return true;
  }

  function amount(value) {
    if (row.unit === "min" && Number(value) === 0)
      return "Never";
    return row.unit ? value + " " + row.unit : String(value);
  }

  function edit() {
    editor.text = row.value;
    editing = true;
    editor.focusInput();
  }

  implicitHeight: layout.implicitHeight + 24
  radius: Theme.radius
  color: current ? Theme.overlay : area.containsMouse ? Theme.surface : "transparent"
  border.width: current ? 2 : 0
  border.color: Theme.accent

  MouseArea {
    id: area

    anchors.fill: parent
    hoverEnabled: true
    onClicked: {
      item.pointed();
      item.activate();
    }
  }

  RowLayout {
    id: layout

    anchors.left: parent.left
    anchors.right: parent.right
    anchors.verticalCenter: parent.verticalCenter
    anchors.leftMargin: 14
    anchors.rightMargin: 14
    spacing: 16

    ColumnLayout {
      Layout.fillWidth: true
      spacing: 3

      Label {
        Layout.fillWidth: true
        text: item.row.label
        font.bold: true
        elide: Text.ElideRight
      }

      Label {
        Layout.fillWidth: true
        visible: !!item.row.help
        text: item.row.help || ""
        color: Theme.subtext
        font.pixelSize: Theme.size(12)
        wrapMode: Text.Wrap
      }

      Label {
        Layout.fillWidth: true
        visible: item.marks.length > 0
        text: item.marks.join("   ")
        color: item.computer && !item.grouped ? Theme.warning : Theme.subtext
        font.pixelSize: Theme.size(11)
      }
    }

    Label {
      visible: item.busy
      text: "\uf110"
      color: Theme.accent
    }

    Toggle {
      visible: item.row.kind === "toggle"
      worded: true
      checked: item.row.value === "yes"
      onToggled: {
        item.pointed();
        item.activate();
      }
    }

    Label {
      visible: item.row.kind === "choice"
      Layout.maximumWidth: item.width * 0.4
      text: item.shownValue + "  \uf0d7"
      color: Theme.accent
      elide: Text.ElideRight
    }

    RowLayout {
      visible: item.row.kind === "slider"
      spacing: 10

      Slider {
        Layout.preferredWidth: Theme.size(180)
        value: item.row.max > item.row.min ? (Number(item.row.value) - item.row.min) / (item.row.max - item.row.min) : 0
        step: item.row.max > item.row.min ? item.row.step / (item.row.max - item.row.min) : 0.05
        onMoved: level => settle.later(item.row.min + Math.round(level * (item.row.max - item.row.min) / item.row.step) * item.row.step)
      }

      Label {
        id: reading

        Layout.preferredWidth: Math.ceil(Math.max(digits.advanceWidth(item.amount(item.row.min)), digits.advanceWidth(item.amount(item.row.max))))
        text: item.amount(item.row.value)
        horizontalAlignment: Text.AlignRight
        font.features: ({
            "tnum": 1
          })

        FontMetrics {
          id: digits

          font: reading.font
        }
      }
    }

    Field {
      id: editor

      readonly property string empty: item.row.kind === "time" ? "07:00" : "The default"

      visible: item.row.kind === "text" || item.row.kind === "time"
      implicitWidth: item.row.kind === "time" ? Theme.size(90) : Math.min(Theme.size(360), item.width / 2)
      placeholder: item.editing ? empty : ""
      text: item.row.value
      onAccepted: {
        item.editing = false;
        Store.set(item.row, text);
      }

      Rectangle {
        anchors.fill: parent
        anchors.margins: 1
        visible: !item.editing
        radius: Math.max(0, Theme.radius - 1)
        color: Theme.surface

        Label {
          anchors.fill: parent
          anchors.leftMargin: 7
          anchors.rightMargin: 7
          text: item.row.value || editor.empty
          color: item.row.value ? Theme.text : Theme.subtext
          elide: Text.ElideRight
        }
      }

      MouseArea {
        anchors.fill: parent
        onClicked: {
          item.pointed();
          item.edit();
        }
      }
    }

    Label {
      visible: item.row.kind === "button"
      text: "Open  \uf35d"
      color: Theme.accent
    }
  }

  Timer {
    id: settle

    property real wanted: 0

    function later(value) {
      wanted = Math.round(value * 100) / 100;
      restart();
    }

    interval: 400
    onTriggered: Store.set(item.row, wanted)
  }
}
