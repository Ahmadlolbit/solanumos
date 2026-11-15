import QtQuick
import QtQuick.Layouts
import qs.common
import qs.common.panels

Label {
  id: root

  property var bar

  text: panel.icon() + " " + Math.round(panel.volume * 100) + "%"
  color: panel.muted ? Theme.subtext : Theme.text

  MouseArea {
    anchors.fill: parent
    acceptedButtons: Qt.LeftButton | Qt.RightButton
    cursorShape: Qt.PointingHandCursor
    onClicked: mouse => {
      if (mouse.button === Qt.RightButton && panel.sink)
        panel.mute(panel.sink);
      else
        popup.toggle();
    }
    onWheel: wheel => {
      const steps = notches.steps(wheel);
      if (steps !== 0)
        panel.setVolume(panel.sink, steps > 0 ? Math.max(panel.volume, Math.min(1, panel.volume + 0.05 * steps)) : Math.max(0, panel.volume + 0.05 * steps));
    }

    Notches {
      id: notches
    }
  }

  BarPopup {
    id: popup

    name: "audio"
    bar: root.bar
    owner: root

    Sound {
      id: panel

      Layout.fillWidth: true
      onDone: Popups.close(popup)
    }
  }
}
