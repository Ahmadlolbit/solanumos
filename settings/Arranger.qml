import QtQuick
import QtQuick.Layouts
import qs.common

Tile {
  id: tile

  property int at: 0
  readonly property bool trying: Arrangement.trying

  signal pointed

  function activate() {
    if (trying) {
      Arrangement.answer(at === 0);
      return true;
    }
    return Arrangement.apply();
  }

  function nudge(key) {
    const way = {
      [Qt.Key_Left]: [-1, 0],
      [Qt.Key_Right]: [1, 0],
      [Qt.Key_Up]: [0, -1],
      [Qt.Key_Down]: [0, 1]
    }[key];
    return !trying && !!way && Arrangement.hop(Arrangement.picked, way[0], way[1]);
  }

  function step(direction) {
    if (trying) {
      if (at + direction < 0 || at + direction > 1)
        return false;
      at += direction;
      return true;
    }
    const count = Arrangement.plan.length;
    if (count < 2)
      return false;
    Arrangement.picked = (Arrangement.picked + direction + count) % count;
    return true;
  }

  title: "Screens"
  hint: trying ? "" : current ? "Drag, or Shift and the arrows, to move · Left and Right pick a screen · Enter applies" : Arrangement.saved ? "Saved for these screens" : "Not saved for these screens yet"
  onTryingChanged: at = 0

  Column {
    width: parent.width
    spacing: 14

    Item {
      id: canvas

      readonly property var boxes: Arrangement.plan.map(Arrangement.box)
      readonly property real minX: boxes.length ? Math.min(...boxes.map(item => item.x)) : 0
      readonly property real minY: boxes.length ? Math.min(...boxes.map(item => item.y)) : 0
      readonly property real wide: boxes.length ? Math.max(...boxes.map(item => item.x + item.w)) - minX : 1
      readonly property real high: boxes.length ? Math.max(...boxes.map(item => item.y + item.h)) - minY : 1
      readonly property real factor: Math.max(0.01, Math.min((width - 32) / Math.max(1, wide), (height - 32) / Math.max(1, high)))
      readonly property real originX: (width - wide * factor) / 2
      readonly property real originY: (height - high * factor) / 2

      width: parent.width
      height: Theme.size(220)

      Rectangle {
        anchors.fill: parent
        radius: Theme.radius
        color: Qt.alpha(Theme.surface, 0.45)
      }

      Repeater {
        model: Arrangement.plan.length

        Rectangle {
          id: shape

          required property int index
          readonly property var item: Arrangement.plan[index] || ({})
          readonly property var spot: canvas.boxes[index] || ({
              x: 0,
              y: 0,
              w: 1,
              h: 1
            })
          readonly property bool picked: index === Arrangement.picked
          property real dragX: 0
          property real dragY: 0

          x: canvas.originX + (spot.x - canvas.minX) * canvas.factor + dragX
          y: canvas.originY + (spot.y - canvas.minY) * canvas.factor + dragY
          z: picked ? 2 : item.on && !item.mirror ? 1 : 0
          width: Math.max(8, spot.w * canvas.factor)
          height: Math.max(8, spot.h * canvas.factor)
          radius: Theme.radius
          color: picked ? Qt.tint(Theme.base, Qt.alpha(Theme.accent, 0.2)) : Theme.base
          opacity: item.on ? 1 : 0.55
          border.width: picked ? 2 : 1
          border.color: picked ? Theme.accent : Theme.overlay

          Column {
            anchors.centerIn: parent
            width: parent.width - 12
            spacing: 2

            Label {
              width: parent.width
              text: shape.item.output || ""
              font.bold: true
              horizontalAlignment: Text.AlignHCenter
              elide: Text.ElideRight
            }

            Label {
              width: parent.width
              text: [(shape.item.size || "").replace("x", " × "), shape.item.main ? "Main" : "", shape.item.on ? "" : "Off", shape.item.mirror ? "Mirrors " + shape.item.mirror : ""].filter(word => word).join(" · ")
              color: Theme.subtext
              font.pixelSize: Theme.size(11)
              horizontalAlignment: Text.AlignHCenter
              elide: Text.ElideRight
            }
          }

          MouseArea {
            property point start

            anchors.fill: parent
            preventStealing: true
            cursorShape: pressed ? Qt.ClosedHandCursor : Qt.OpenHandCursor
            onPressed: mouse => {
              start = mapToItem(canvas, mouse.x, mouse.y);
              tile.pointed();
              Arrangement.picked = shape.index;
            }
            onPositionChanged: mouse => {
              const now = mapToItem(canvas, mouse.x, mouse.y);
              shape.dragX = now.x - start.x;
              shape.dragY = now.y - start.y;
            }
            onReleased: {
              Arrangement.drop(shape.index, shape.spot.x + shape.dragX / canvas.factor, shape.spot.y + shape.dragY / canvas.factor, 12 / canvas.factor);
              shape.dragX = 0;
              shape.dragY = 0;
            }
            onCanceled: {
              shape.dragX = 0;
              shape.dragY = 0;
            }
          }
        }
      }
    }

    RowLayout {
      width: parent.width
      spacing: 10
      visible: !tile.trying

      Label {
        Layout.fillWidth: true
        text: Arrangement.changed ? "Changed, not applied yet" : Arrangement.saved ? "As saved for these screens" : "As Hyprland set them up"
        color: Theme.subtext
        elide: Text.ElideRight
      }

      Button {
        visible: Arrangement.changed
        label: "Undo"
        onClicked: {
          tile.pointed();
          Arrangement.undo();
        }
      }

      Button {
        visible: Arrangement.changed
        label: "Apply"
        primary: true
        chosen: tile.current
        onClicked: {
          tile.pointed();
          Arrangement.apply();
        }
      }
    }

    RowLayout {
      width: parent.width
      spacing: 10
      visible: tile.trying

      Label {
        Layout.fillWidth: true
        text: "Keep this layout? The old one comes back in " + Arrangement.left + " s"
        font.bold: true
        wrapMode: Text.Wrap
      }

      Button {
        label: "Keep"
        primary: true
        chosen: tile.current && tile.at === 0
        onClicked: {
          tile.pointed();
          tile.at = 0;
          Arrangement.answer(true);
        }
      }

      Button {
        label: "Put back"
        chosen: tile.current && tile.at === 1
        onClicked: {
          tile.pointed();
          tile.at = 1;
          Arrangement.answer(false);
        }
      }
    }
  }
}
