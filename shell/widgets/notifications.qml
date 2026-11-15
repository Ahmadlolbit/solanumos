import QtQuick
import QtQuick.Layouts
import Quickshell
import qs.common
import qs.common.panels

Label {
  id: root

  property var bar
  readonly property var history: Notifications.history
  readonly property bool quiet: Notifications.quiet
  readonly property var shownItems: Notifications.shownItems
  readonly property int unread: Notifications.unread

  function remove(event, item) {
    Keyboard.used = true;
    if (event.modifiers & Qt.ShiftModifier)
      Notifications.clear();
    else if (item)
      Notifications.dismiss(item.id);
    else
      return false;
    return true;
  }

  text: (quiet ? "\u{f009b}" : "\u{f009a}") + (unread > 0 ? " " + unread : "")
  color: unread > 0 && !quiet ? Theme.accent : Theme.subtext

  MouseArea {
    anchors.fill: parent
    cursorShape: Qt.PointingHandCursor
    onClicked: popup.toggle()
  }

  BarPopup {
    id: popup

    name: "notifications"
    bar: root.bar
    owner: root
    onVisibleChanged: {
      if (visible)
        Notifications.open();
    }

    ColumnLayout {
      Layout.fillWidth: true
      spacing: 10
      Keys.onDeletePressed: event => event.accepted = root.remove(event, null)

      RowLayout {
        Layout.fillWidth: true
        Layout.preferredWidth: Theme.size(340)
        spacing: 10

        Label {
          Layout.fillWidth: true
          text: "Notifications"
          font.bold: true
        }

        Label {
          text: "Do not disturb"
          color: Theme.subtext
        }

        Toggle {
          worded: true
          checked: root.quiet
          onToggled: Notifications.toggleQuiet()
        }
      }

      Label {
        visible: root.shownItems.length === 0
        text: "Nothing yet"
        color: Theme.subtext
      }

      Rows {
        id: list

        Layout.fillWidth: true
        Layout.preferredHeight: Math.min(contentHeight, 360)
        model: ScriptModel {
          values: root.shownItems
          objectProp: "id"
        }

        delegate: Rectangle {
          id: item

          required property var modelData

          width: list.width
          height: lines.implicitHeight + 8
          radius: Theme.radius
          color: "transparent"
          border.width: activeFocus && Keyboard.used ? Theme.focusBorder : 0
          border.color: Theme.accent
          Keys.onDeletePressed: event => event.accepted = root.remove(event, item.modelData)

          ColumnLayout {
            id: lines

            anchors.left: parent.left
            anchors.right: parent.right
            anchors.verticalCenter: parent.verticalCenter
            anchors.margins: 8
            spacing: 2

            Label {
              Layout.fillWidth: true
              text: item.modelData.app
              color: item.modelData.critical ? Theme.critical : Theme.subtext
              elide: Text.ElideRight
            }

            Label {
              Layout.fillWidth: true
              text: item.modelData.summary
              font.bold: true
              elide: Text.ElideRight
            }

            Label {
              Layout.fillWidth: true
              visible: text !== ""
              text: item.modelData.body
              wrapMode: Text.WordWrap
              maximumLineCount: 3
              elide: Text.ElideRight
            }
          }
        }
      }

      Choice {
        Layout.fillWidth: true
        visible: root.shownItems.length > 0
        icon: "\u{f1f8}"
        label: "Clear"
        onClicked: Notifications.clear()
      }

      Choice {
        Layout.fillWidth: true
        visible: root.history.length > 0
        icon: "\u{f1da}"
        label: "Show the last one"
        onClicked: {
          Popups.close(popup);
          Quickshell.execDetached(["makoctl", "restore"]);
        }
      }
    }
  }
}
