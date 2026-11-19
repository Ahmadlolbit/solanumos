import QtQuick

Rectangle {
  id: keys

  property Item target
  property color ink: "white"
  property color paper: "black"
  property color accent: "white"
  property string family: "monospace"
  property real factor: 1
  property bool shifted: false
  property bool signs: false
  readonly property var letters: ["1234567890", "qwertyuiop", "asdfghjkl", "zxcvbnm"]
  readonly property var marks: ["1234567890", "!@#$%^&*()", "-_=+[]{};:", "'\",.<>/?\\|"]
  readonly property int unit: Math.round(48 * factor)
  readonly property int gap: Math.round(6 * factor)

  signal typed
  signal entered

  function type(text) {
    target.insert(target.cursorPosition, shifted ? text.toUpperCase() : text);
    shifted = false;
    typed();
  }

  function erase() {
    if (target.cursorPosition > 0)
      target.remove(target.cursorPosition - 1, target.cursorPosition);
    typed();
  }

  function enter() {
    entered();
  }

  implicitWidth: rows.implicitWidth + 2 * gap
  implicitHeight: rows.implicitHeight + 2 * gap
  color: Qt.alpha(paper, 0.9)
  border.color: Qt.alpha(ink, 0.3)

  component Key: Rectangle {
    id: key

    property string label
    property real span: 1
    property bool lit: false

    signal hit

    width: keys.unit * span + (Math.ceil(span) - 1) * keys.gap
    height: keys.unit
    color: area.pressed || lit ? keys.accent : Qt.alpha(keys.ink, 0.12)
    border.color: Qt.alpha(keys.ink, 0.3)

    Text {
      anchors.centerIn: parent
      text: key.label
      color: area.pressed || key.lit ? keys.paper : keys.ink
      font.family: keys.family
      font.pixelSize: Math.round(18 * keys.factor)
    }

    MouseArea {
      id: area

      anchors.fill: parent
      cursorShape: Qt.PointingHandCursor
      onClicked: key.hit()
    }
  }

  Column {
    id: rows

    anchors.centerIn: parent
    spacing: keys.gap

    Repeater {
      model: keys.signs ? keys.marks : keys.letters

      Row {
        required property string modelData

        anchors.horizontalCenter: parent.horizontalCenter
        spacing: keys.gap

        Repeater {
          model: parent.modelData.split("")

          Key {
            required property string modelData

            label: keys.shifted ? modelData.toUpperCase() : modelData
            onHit: keys.type(modelData)
          }
        }
      }
    }

    Row {
      anchors.horizontalCenter: parent.horizontalCenter
      spacing: keys.gap

      Key {
        label: "Shift"
        span: 2
        lit: keys.shifted
        visible: !keys.signs
        onHit: keys.shifted = !keys.shifted
      }

      Key {
        label: keys.signs ? "abc" : "?123"
        span: 2
        onHit: keys.signs = !keys.signs
      }

      Key {
        label: "Space"
        span: 4
        onHit: keys.type(" ")
      }

      Key {
        label: "Delete"
        span: 2
        onHit: keys.erase()
      }

      Key {
        label: "Enter"
        span: 2
        onHit: keys.enter()
      }
    }
  }
}
