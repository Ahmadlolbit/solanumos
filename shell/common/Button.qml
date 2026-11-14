import QtQuick

Rectangle {
  id: root

  property string label: ""
  property string glyph: ""
  property bool primary: false
  readonly property bool filled: primary !== (area.containsMouse || activeFocus)
  readonly property color ink: primary ? Theme.accent : Theme.text

  signal clicked

  implicitWidth: content.implicitWidth + 32
  implicitHeight: Theme.size(30)
  radius: Theme.radius
  color: filled ? ink : "transparent"
  border.width: Theme.focusBorder
  border.color: activeFocus ? Theme.accent : ink
  activeFocusOnTab: true
  Keys.onPressed: event => {
    const pressed = [Qt.Key_Return, Qt.Key_Enter, Qt.Key_Space].includes(event.key);
    if (pressed && !event.isAutoRepeat)
      root.clicked();
    event.accepted = pressed;
  }

  Row {
    id: content

    anchors.centerIn: parent
    spacing: 8

    Label {
      visible: root.glyph !== ""
      text: root.glyph
      color: root.filled ? Theme.base : root.ink
    }

    Label {
      text: root.label
      color: root.filled ? Theme.base : root.ink
    }
  }

  MouseArea {
    id: area

    anchors.fill: parent
    hoverEnabled: true
    cursorShape: Qt.PointingHandCursor
    onClicked: root.clicked()
  }
}
