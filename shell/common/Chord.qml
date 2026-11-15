import QtQuick

Row {
  id: root

  property string keys: ""

  spacing: 4

  Repeater {
    model: root.keys.split(" + ")

    Row {
      id: part

      required property string modelData
      required property int index

      spacing: 4

      Label {
        anchors.verticalCenter: parent.verticalCenter
        visible: part.index > 0
        text: "+"
        color: Theme.subtext
      }

      Keycap {
        anchors.verticalCenter: parent.verticalCenter
        key: part.modelData
      }
    }
  }
}
