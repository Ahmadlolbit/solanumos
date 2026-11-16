import QtQuick
import QtQuick.Layouts
import Quickshell
import qs.common

Label {
  id: root

  property var bar
  property var day: new Date()
  readonly property var month: new Date(day.getFullYear(), day.getMonth(), 1)
  readonly property var moves: ({
      [Qt.Key_Left]: -1,
      [Qt.Key_Right]: 1,
      [Qt.Key_Up]: -7,
      [Qt.Key_Down]: 7
    })

  function days() {
    const offset = (month.getDay() - Qt.locale().firstDayOfWeek + 7) % 7;
    return Array.from({
      length: 42
    }, (_, index) => new Date(month.getFullYear(), month.getMonth(), 1 - offset + index));
  }

  function move(count) {
    day = new Date(day.getFullYear(), day.getMonth(), day.getDate() + count);
  }

  function shift(months) {
    const last = new Date(day.getFullYear(), day.getMonth() + months + 1, 0).getDate();
    day = new Date(day.getFullYear(), day.getMonth() + months, Math.min(day.getDate(), last));
  }

  function reset() {
    day = new Date(clock.date.getFullYear(), clock.date.getMonth(), clock.date.getDate());
  }

  function key(event) {
    Keyboard.used = true;
    if (moves[event.key] !== undefined)
      move(moves[event.key]);
    else if (event.key === Qt.Key_PageUp || event.key === Qt.Key_PageDown)
      shift(event.key === Qt.Key_PageUp ? -1 : 1);
    else if (event.key === Qt.Key_Home)
      reset();
    else
      return false;
    return true;
  }

  text: Qt.formatDateTime(clock.date, root.bar && root.bar.vertical ? (Config.clock === "12h" ? "h\nmm" : "HH\nmm") : Config.clockFormat || (Config.clock === "12h" ? "ddd d MMM   h:mm AP" : "ddd d MMM   HH:mm"))
  horizontalAlignment: Text.AlignHCenter
  font.bold: true

  SystemClock {
    id: clock

    precision: SystemClock.Minutes
  }

  MouseArea {
    anchors.fill: parent
    cursorShape: Qt.PointingHandCursor
    onClicked: popup.toggle()
  }

  BarPopup {
    id: popup

    name: "calendar"
    bar: root.bar
    owner: root
    onVisibleChanged: {
      if (visible)
        root.reset();
    }

    RowLayout {
      Layout.fillWidth: true
      spacing: 10

      Label {
        text: "\u{f0141}"

        MouseArea {
          anchors.fill: parent
          cursorShape: Qt.PointingHandCursor
          onClicked: root.shift(-1)
        }
      }

      Label {
        Layout.fillWidth: true
        horizontalAlignment: Text.AlignHCenter
        text: Qt.locale().standaloneMonthName(root.month.getMonth()) + " " + root.month.getFullYear()
        font.bold: true

        MouseArea {
          anchors.fill: parent
          cursorShape: Qt.PointingHandCursor
          onClicked: root.reset()
        }
      }

      Label {
        text: "\u{f0142}"

        MouseArea {
          anchors.fill: parent
          cursorShape: Qt.PointingHandCursor
          onClicked: root.shift(1)
        }
      }
    }

    GridLayout {
      id: grid

      columns: 7
      columnSpacing: 2 * Theme.focusBorder
      rowSpacing: 2 * Theme.focusBorder
      activeFocusOnTab: true
      Keys.onPressed: event => event.accepted = root.key(event)

      Repeater {
        model: 7

        Label {
          required property int index

          Layout.preferredWidth: Theme.size(34)
          horizontalAlignment: Text.AlignHCenter
          text: Qt.locale().dayName((Qt.locale().firstDayOfWeek + index) % 7, Locale.ShortFormat)
          color: Theme.subtext
        }
      }

      Repeater {
        model: root.days()

        Rectangle {
          id: cell

          required property var modelData
          readonly property bool today: modelData.toDateString() === clock.date.toDateString()

          Layout.preferredWidth: Theme.size(34)
          Layout.preferredHeight: Theme.size(28)
          radius: Theme.radius
          color: today ? Theme.accent : "transparent"

          Label {
            anchors.centerIn: parent
            text: cell.modelData.getDate()
            color: cell.today ? Theme.base : cell.modelData.getMonth() === root.month.getMonth() ? Theme.text : Theme.subtext
          }

          Rectangle {
            anchors.fill: parent
            anchors.margins: -2 * Theme.focusBorder
            visible: grid.activeFocus && Keyboard.used && cell.modelData.toDateString() === root.day.toDateString()
            radius: Theme.radius ? Theme.radius + 2 * Theme.focusBorder : 0
            color: "transparent"
            border.width: Theme.focusBorder
            border.color: Theme.accent
          }
        }
      }
    }
  }
}
