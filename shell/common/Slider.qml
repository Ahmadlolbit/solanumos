import QtQuick

Item {
  id: root

  property real value: 0
  property real step: 0.05
  readonly property real shown: Math.max(0, Math.min(1, value))

  signal moved(real level)

  function nudge(steps) {
    moved(steps > 0 ? Math.max(value, Math.min(1, value + step * steps)) : Math.max(0, value + step * steps));
  }

  function key(event) {
    Keyboard.used = true;
    if (event.key === Qt.Key_Home || event.key === Qt.Key_End)
      moved(event.key === Qt.Key_Home ? 0 : 1);
    else if (event.key === Qt.Key_Left || event.key === Qt.Key_Right)
      nudge(event.key === Qt.Key_Left ? -1 : 1);
    else
      return false;
    return true;
  }

  implicitWidth: 200
  implicitHeight: 20
  activeFocusOnTab: true
  Keys.onPressed: event => event.accepted = root.key(event)

  Rectangle {
    id: track

    anchors.left: parent.left
    anchors.right: parent.right
    anchors.verticalCenter: parent.verticalCenter
    height: 6
    radius: Theme.radius ? height / 2 : 0
    color: Theme.overlay

    Rectangle {
      width: track.width * root.shown
      height: parent.height
      radius: parent.radius
      color: Theme.accent
    }
  }

  Rectangle {
    id: knob

    x: (root.width - width) * root.shown
    anchors.verticalCenter: parent.verticalCenter
    width: 14
    height: 14
    radius: Theme.radius ? width / 2 : 0
    color: Theme.text
  }

  Rectangle {
    anchors.fill: knob
    anchors.margins: -2 * Theme.focusBorder
    visible: root.activeFocus && Keyboard.used
    radius: Theme.radius ? width / 2 : 0
    color: "transparent"
    border.width: Theme.focusBorder
    border.color: Theme.accent
  }

  MouseArea {
    function place(x) {
      root.moved(Math.max(0, Math.min(1, x / width)));
    }

    anchors.fill: parent
    cursorShape: Qt.PointingHandCursor
    onPressed: mouse => place(mouse.x)
    onPositionChanged: mouse => place(mouse.x)
    onWheel: wheel => {
      const steps = notches.steps(wheel);
      if (steps !== 0)
        root.nudge(steps);
    }

    Notches {
      id: notches
    }
  }
}
