import QtQuick

Behavior {
  id: root

  NumberAnimation {
    duration: root.targetValue ? Theme.fadeIn : Theme.fadeOut
    easing.type: root.targetValue ? Easing.OutCubic : Easing.Linear
  }
}
