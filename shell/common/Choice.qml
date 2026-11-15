import QtQuick
import QtQuick.Layouts

Rectangle {
  id: root

  property string icon: ""
  property string label: ""
  property string detail: ""
  property string action: ""
  property bool selected: false
  readonly property bool offered: action !== "" && (area.containsMouse || activeFocus)

  signal clicked
  signal actionClicked

  function key(event) {
    const pressed = [Qt.Key_Return, Qt.Key_Enter, Qt.Key_Space].includes(event.key);
    const acted = event.key === Qt.Key_Delete && root.action !== "";
    if (pressed && !event.isAutoRepeat)
      root.clicked();
    else if (acted && !event.isAutoRepeat)
      root.actionClicked();
    return pressed || acted;
  }

  implicitWidth: row.implicitWidth + 16
  implicitHeight: Theme.size(30)
  radius: Theme.radius
  color: selected ? Theme.overlay : area.containsMouse ? Theme.surface : "transparent"
  border.width: activeFocus ? Theme.focusBorder : 0
  border.color: Theme.accent
  activeFocusOnTab: true
  Keys.onPressed: event => event.accepted = root.key(event)

  MouseArea {
    id: area

    anchors.fill: parent
    hoverEnabled: true
    cursorShape: Qt.PointingHandCursor
    onClicked: root.clicked()
  }

  RowLayout {
    id: row

    anchors.fill: parent
    anchors.leftMargin: 8
    anchors.rightMargin: 8
    spacing: 10

    Label {
      visible: root.icon !== ""
      text: root.icon
      color: root.selected ? Theme.accent : Theme.text
    }

    Label {
      Layout.fillWidth: true
      text: root.label
      elide: Text.ElideRight
    }

    Label {
      visible: root.detail !== "" && !root.offered
      text: root.detail
      color: Theme.subtext
    }

    Label {
      visible: root.offered
      text: root.action
      color: Theme.accent

      MouseArea {
        anchors.fill: parent
        cursorShape: Qt.PointingHandCursor
        onClicked: root.actionClicked()
      }
    }
  }
}
