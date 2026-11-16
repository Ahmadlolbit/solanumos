import QtQuick

Item {
  id: root

  property bool checked: false
  property bool worded: false

  signal toggled

  implicitWidth: (worded ? word.width + 8 : 0) + track.width
  implicitHeight: 20
  activeFocusOnTab: true
  Keys.onPressed: event => {
    Keyboard.used = true;
    const pressed = [Qt.Key_Space, Qt.Key_Return, Qt.Key_Enter].includes(event.key);
    if (pressed && !event.isAutoRepeat)
      root.toggled();
    event.accepted = pressed;
  }

  TextMetrics {
    id: widest

    font: word.font
    text: "Off"
  }

  Label {
    id: word

    anchors.right: track.left
    anchors.rightMargin: 8
    anchors.verticalCenter: parent.verticalCenter
    width: Math.ceil(widest.advanceWidth)
    visible: root.worded
    text: root.checked ? "On" : "Off"
    color: Theme.subtext
    horizontalAlignment: Text.AlignRight
  }

  Rectangle {
    id: track

    anchors.right: parent.right
    anchors.verticalCenter: parent.verticalCenter
    width: 40
    height: 20
    radius: Theme.radius ? height / 2 : 0
    color: root.checked ? Theme.accent : "transparent"
    border.width: root.checked ? 0 : 1
    border.color: Theme.subtext

    Rectangle {
      x: root.checked ? track.width - width - 3 : 3
      anchors.verticalCenter: parent.verticalCenter
      width: track.height - 6
      height: width
      radius: Theme.radius ? width / 2 : 0
      color: root.checked ? Theme.base : Theme.subtext

      Behavior on x {
        NumberAnimation {
          duration: Theme.quick
        }
      }
    }
  }

  Rectangle {
    anchors.fill: track
    anchors.margins: -2 * Theme.focusBorder
    visible: root.activeFocus && Keyboard.used
    radius: Theme.radius ? height / 2 : 0
    color: "transparent"
    border.width: Theme.focusBorder
    border.color: Theme.accent
  }

  MouseArea {
    anchors.fill: parent
    cursorShape: Qt.PointingHandCursor
    onClicked: root.toggled()
  }
}
